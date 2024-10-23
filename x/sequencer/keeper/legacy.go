package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// unbond unbonds a sequencer
// if jail is true, the sequencer is jailed as well (cannot be bonded again)
// bonded tokens are refunded by default, unless jail is true
func (k Keeper) unbondLegacy(ctx sdk.Context, seqAddr string, jail bool) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrSequencerNotFound
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

		// if we unbond the proposer, remove it
		// the caller should rotate the proposer
		if k.isProposer(ctx, seq.RollappId, seqAddr) {
			k.removeProposer(ctx, seq.RollappId)
		}

		// if we unbond the next proposer, we're in the middle of rotation
		// the caller should clean the rotation state
		if k.isNextProposer(ctx, seq.RollappId, seqAddr) {
			k.removeNextProposer(ctx, seq.RollappId)
		}
	}
	// in case the sequencer is currently reducing its bond, then we need to remove it from the decreasing bond queue
	// all the tokens are returned, so we don't need to reduce the bond anymore
	if bondReductionIDs := k.getBondReductionIDsBySequencer(ctx, seq.Address); len(bondReductionIDs) > 0 {
		for _, bondReductionID := range bondReductionIDs {
			bondReduction, found := k.GetBondReduction(ctx, bondReductionID)
			if found {
				k.removeBondReduction(ctx, bondReductionID, bondReduction)
			}
		}
	}

	if jail {
		seq.Jailed = true
	}
	// set the unbonding height and time, if not already set.
	// to avoid leaving unbonded sequencer in the store with no unbond height or time
	if seq.UnbondRequestHeight == 0 {
		seq.UnbondRequestHeight = ctx.BlockHeight()
	}
	if seq.UnbondTime.IsZero() {
		seq.UnbondTime = ctx.BlockTime()
	}

	// update the sequencer in store
	seq.Status = types.Unbonded
	k.UpdateSequencer(ctx, &seq, oldStatus)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
		),
	)

	return nil
}
