package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// startUnbondingPeriodForSequencer sets the sequencer to unbonding status
// can be called after notice period or directly if notice period is not required
// caller is responsible for updating the proposer for the rollapp if needed
func (k Keeper) startUnbondingPeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) time.Time {
	completionTime := ctx.BlockTime().Add(k.UnbondingTime(ctx))
	seq.UnbondTime = completionTime

	seq.Status = types.Unbonding
	k.UpdateSequencer(ctx, *seq, types.Bonded)
	k.AddSequencerToUnbondingQueue(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonding,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.SequencerAddress),
			sdk.NewAttribute(types.AttributeKeyBond, seq.Tokens.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.String()),
		),
	)

	return completionTime
}

// UnbondAllMatureSequencers unbonds all the mature unbonding sequencers that
// have finished their unbonding period.
func (k Keeper) UnbondAllMatureSequencers(ctx sdk.Context, currTime time.Time) {
	sequencers := k.GetMatureUnbondingSequencers(ctx, currTime)
	for _, seq := range sequencers {
		wrapFn := func(ctx sdk.Context) error {
			return k.unbondSequencer(ctx, seq.SequencerAddress)
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

func (k Keeper) unbondSequencerAndBurn(ctx sdk.Context, seqAddr string) (*types.Sequencer, error) {
	return k.unbondSequencerBurnOrRefund(ctx, seqAddr, true)
}

func (k Keeper) unbondSequencer(ctx sdk.Context, seqAddr string) error {
	_, err := k.unbondSequencerBurnOrRefund(ctx, seqAddr, false)
	return err
}

func (k Keeper) unbondSequencerBurnOrRefund(ctx sdk.Context, seqAddr string, burnBond bool) (*types.Sequencer, error) {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is already unbonded",
		)
	}
	// keep the old status for updating the sequencer
	oldStatus := seq.Status

	// handle bond
	seqTokens := seq.Tokens
	if !seqTokens.Empty() {
		if burnBond {
			err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, seqTokens)
			if err != nil {
				return nil, err
			}
		} else { //refund
			seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
			if err != nil {
				return nil, err
			}

			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seqTokens)
			if err != nil {
				return nil, err
			}
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.Address)
	}

	// remove from queue if unbonding
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

	// remove from notice period queue if needed
	if seq.IsNoticePeriodInProgress() {
		k.removeNoticePeriodSequencer(ctx, seq)
	}

	// if the slashed sequencer is the proposer, remove it
	// the caller should rotate the proposer
	if k.isProposer(ctx, seq.RollappId, seqAddr) {
		k.removeProposer(ctx, seq.RollappId)
	}

	// if we slash the next proposer, we're in the middle of rotation
	// instead of removing the next proposer, we set it to empty, and the chain will halt
	if k.isNextProposer(ctx, seq.RollappId, seqAddr) {
		k.setNextProposer(ctx, seq.RollappId, NO_SEQUENCER_AVAILABLE)
	}

	// update the sequencer in store
	seq.Status = types.Unbonded
	seq.Tokens = sdk.Coins{}
	k.UpdateSequencer(ctx, seq, oldStatus)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyBond, seqTokens.String()),
		),
	)

	return &seq, nil
}

func (k Keeper) completeBondReduction(ctx sdk.Context, reduction types.BondReduction) error {
	seq, found := k.GetSequencer(ctx, reduction.SequencerAddress)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Tokens.IsAllLT(sdk.NewCoins(reduction.DecreaseBondAmount)) {
		return errorsmod.Wrapf(
			types.ErrInsufficientBond,
			"sequencer does not have enough bond to reduce insufficient bond: got %s, reducing by %s",
			seq.Tokens.String(),
			reduction.DecreaseBondAmount.String(),
		)
	}
	newBalance := seq.Tokens.Sub(reduction.DecreaseBondAmount)
	// in case between unbonding queue and now, the minbond value is increased,
	// handle it by only returning upto minBond amount and not all
	minBond := k.GetParams(ctx).MinBond
	if newBalance.IsAllLT(sdk.NewCoins(minBond)) {
		diff := minBond.SubAmount(newBalance.AmountOf(minBond.Denom))
		reduction.DecreaseBondAmount = reduction.DecreaseBondAmount.Sub(diff)
	}
	seqAddr := sdk.MustAccAddressFromBech32(reduction.SequencerAddress)
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAddr, sdk.NewCoins(reduction.DecreaseBondAmount))
	if err != nil {
		return err
	}

	seq.Tokens = seq.Tokens.Sub(reduction.DecreaseBondAmount)
	k.SetSequencer(ctx, seq)
	k.removeDecreasingBondQueue(ctx, reduction)

	return nil
}
