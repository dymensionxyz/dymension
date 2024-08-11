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
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
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
			return k.unbondSequencer(ctx, seq.Address)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("unbond sequencer", "error", err, "sequencer", seq.Address)
			continue
		}
	}
}

// InstantUnbondAllSequencers unbonds all sequencers for a rollapp
// This is called when the rollapp is frozen
func (k Keeper) InstantUnbondAllSequencers(ctx sdk.Context, rollappID string) error {
	// unbond all bonded/unbonding sequencers
	bonded := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Bonded)
	unbonding := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Unbonding)
	for _, sequencer := range append(bonded, unbonding...) {
		err := k.unbondSequencer(ctx, sequencer.Address)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) reduceSequencerBond(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coins, burn bool) error {
	if amt.IsZero() {
		return nil
	}
	if !seq.Tokens.IsAllGTE(amt) {
		return errorsmod.Wrapf(
			types.ErrInsufficientBond,
			"insufficient bond for sequencer: %s", seq.Address,
		)
	}
	if burn {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
		if err != nil {
			return err
		}
	} else {
		// refund
		seqAcc := sdk.MustAccAddressFromBech32(seq.Address)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, amt)
		if err != nil {
			return err
		}
	}

	seq.Tokens = seq.Tokens.Sub(amt...)
	return nil
}

func (k Keeper) unbondSequencerAndJail(ctx sdk.Context, seqAddr string) error {
	return k.unbond(ctx, seqAddr, true)
}

func (k Keeper) unbondSequencer(ctx sdk.Context, seqAddr string) error {
	return k.unbond(ctx, seqAddr, false)
}

func (k Keeper) unbond(ctx sdk.Context, seqAddr string, jail bool) error {
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
	// keep the old status for updating the sequencer
	oldStatus := seq.Status

	// handle bond: tokens refunded by default, unless jail is true
	err := k.reduceSequencerBond(ctx, &seq, seq.Tokens, jail)
	if err != nil {
		return errorsmod.Wrap(err, "remove sequencer bond")
	}

	/* ------------------------------ store cleanup ----------------------------- */
	// remove from queue if unbonding
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	} else {
		// remove from notice period queue if needed
		if seq.IsNoticePeriodInProgress() {
			k.removeNoticePeriodSequencer(ctx, seq)
		}

		// if we unbond the the proposer, remove it
		// the caller should rotate the proposer
		if k.isProposer(ctx, seq.RollappId, seqAddr) {
			k.removeProposer(ctx, seq.RollappId)
		}

		// if we unbond the next proposer, we're in the middle of rotation
		// instead of removing the next proposer, we set it to empty, and the chain will halt
		// FIXME: review again
		if k.isNextProposer(ctx, seq.RollappId, seqAddr) {
			k.setNextProposer(ctx, seq.RollappId, NO_SEQUENCER_AVAILABLE)
		}
	}
	// in case the sequencer is currently reducing its bond, then we need to remove it from the decreasing bond queue
	// all the tokens are returned, so we don't need to reduce the bond anymore
	if bondReductions := k.getSequencerDecreasingBonds(ctx, seq.Address); len(bondReductions) > 0 {
		for _, bondReduce := range bondReductions {
			k.removeDecreasingBondQueue(ctx, bondReduce)
		}
	}

	if jail {
		seq.Jailed = true
	}
	// set the unbonding height and time, if not already set
	if seq.UnbondRequestHeight == 0 {
		seq.UnbondRequestHeight = ctx.BlockHeight()
	}
	if seq.UnbondTime.IsZero() {
		seq.UnbondTime = ctx.BlockTime()
	}

	// update the sequencer in store
	seq.Status = types.Unbonded
	k.UpdateSequencer(ctx, seq, oldStatus)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
		),
	)

	return nil
}
