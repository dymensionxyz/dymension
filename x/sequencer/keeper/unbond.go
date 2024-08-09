package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UnbondAllMatureSequencers unbonds all the mature unbonding sequencers that
// have finished their unbonding period.
func (k Keeper) UnbondAllMatureSequencers(ctx sdk.Context, currTime time.Time) {
	sequencers := k.GetMatureUnbondingSequencers(ctx, currTime)
	for _, seq := range sequencers {
		wrapFn := func(ctx sdk.Context) error {
			return k.unbondUnbondingSequencer(ctx, seq.Address)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("unbond sequencer", "error", err, "sequencer", seq.Address)
			continue
		}
	}
}

func (k Keeper) HandleBondReduction(ctx sdk.Context, currTime time.Time) {
	unbondings := k.GetMatureDecreasingBondSequencers(ctx, currTime)
	for _, unbonding := range unbondings {
		wrapFn := func(ctx sdk.Context) error {
			return k.completeBondReduction(ctx, unbonding)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("reducing sequencer bond", "error", err, "sequencer", unbonding.SequencerAddress)
			continue
		}
	}
}

func (k Keeper) forceUnbondSequencer(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is already unbonded",
		)
	}

	oldStatus := seq.Status

	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		seqAcc, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.Address)
	}

	// set the status to unbonded and remove from the unbonding queue if needed
	seq.Status = types.Unbonded
	seq.Proposer = false
	seq.Tokens = sdk.Coins{}

	k.UpdateSequencer(ctx, seq, oldStatus)

	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	} else {
		// in case the sequencer is currently reducing its bond, then we need to remove it from the decreasing bond queue
		// all the tokens are returned, so we don't need to reduce the bond anymore
		if bondReductions := k.getSequencerDecreasingBonds(ctx, seq.Address); len(bondReductions) > 0 {
			for _, bondReduce := range bondReductions {
				k.removeDecreasingBondQueue(ctx, bondReduce)
			}
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	return nil
}

// unbondUnbondingSequencer unbonds a sequencer that currently unbonding
func (k Keeper) unbondUnbondingSequencer(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status != types.Unbonding {
		return errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not unbonding: got %s",
			seq.Status.String(),
		)
	}
	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		seqAcc, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.Address)
	}

	// set the status to unbonded and remove from the unbonding queue
	seq.Status = types.Unbonded
	seq.Tokens = sdk.Coins{}

	k.UpdateSequencer(ctx, seq, types.Unbonding)
	k.removeUnbondingSequencer(ctx, seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	return nil
}

func (k Keeper) completeBondReduction(ctx sdk.Context, reduction types.BondReduction) error {
	seq, found := k.GetSequencer(ctx, reduction.SequencerAddress)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Tokens.IsAllLT(sdk.NewCoins(reduction.UnbondAmount)) {
		return errorsmod.Wrapf(
			types.ErrInsufficientBond,
			"sequencer does not have enough bond to reduce insufficient bond: got %s, reducing by %s",
			seq.Tokens.String(),
			reduction.UnbondAmount.String(),
		)
	}
	newBalance := seq.Tokens.Sub(reduction.UnbondAmount)
	// in case between unbonding queue and now, the minbond value is increased,
	// handle it by only returning upto minBond amount and not all
	minBond := k.GetParams(ctx).MinBond
	if newBalance.IsAllLT(sdk.NewCoins(minBond)) {
		diff := minBond.SubAmount(newBalance.AmountOf(minBond.Denom))
		reduction.UnbondAmount = reduction.UnbondAmount.Sub(diff)
	}
	seqAddr := sdk.MustAccAddressFromBech32(reduction.SequencerAddress)
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAddr, sdk.NewCoins(reduction.UnbondAmount))
	if err != nil {
		return err
	}

	seq.Tokens = seq.Tokens.Sub(reduction.UnbondAmount)
	k.SetSequencer(ctx, seq)
	k.removeDecreasingBondQueue(ctx, reduction)

	return nil
}
