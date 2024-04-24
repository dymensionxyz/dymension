package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// UnbondAllMatureSequencers unbonds all the mature unbonding sequencers that
// have finished their unbonding period.
func (k Keeper) UnbondAllMatureSequencers(ctx sdk.Context, currTime time.Time) {
	sequencers := k.GetMatureUnbondingSequencers(ctx, currTime)
	for _, seq := range sequencers {
		wrapFn := func(ctx sdk.Context) error {
			return k.unbondUnbondingSequencer(ctx, seq.SequencerAddress)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("unbond sequencer", "error", err, "sequencer", seq.SequencerAddress)
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
		seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.SequencerAddress)
	}

	// set the status to unbonded and remove from the unbonding queue if needed
	seq.Status = types.Unbonded
	seq.Proposer = false
	seq.Tokens = sdk.Coins{}

	k.UpdateSequencer(ctx, seq, oldStatus)

	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
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
		seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.SequencerAddress)
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
