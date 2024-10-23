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
		if k.isProposerLeg(ctx, seq.RollappId, seqAddr) {
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
	k.UpdateSequencerLeg(ctx, &seq, oldStatus)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonded,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
		),
	)

	return nil
}

// GetSequencer returns a sequencer from its index
func (k Keeper) GetSequencerLegacy(ctx sdk.Context, sequencerAddress string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(
		sequencerAddress,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// MustGetSequencerLeg returns a sequencer from its index
// It will panic if the sequencer is not found
func (k Keeper) MustGetSequencerLeg(ctx sdk.Context, sequencerAddress string) types.Sequencer {
	seq, found := k.GetSequencer(ctx, sequencerAddress)
	if !found {
		panic("sequencer not found")
	}
	return seq
}

/* ------------------------- proposer/next proposer ------------------------- */

// GetProposerLegacy returns the proposer for a rollapp
func (k Keeper) GetProposerLegacy(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ProposerByRollappKey(rollappId))
	if len(b) == 0 || string(b) == types.SentinelSeqAddr {
		return val, false
	}

	return k.GetSequencer(ctx, string(b))
}

func (k Keeper) removeProposer(ctx sdk.Context, rollappId string) {
	k.SetProposer(ctx, rollappId, types.SentinelSeqAddr)
}

func (k Keeper) isProposerLeg(ctx sdk.Context, seq types.Sequencer) bool {
	proposer, ok := k.GetProposerLegacy(ctx, seq.RollappId)
	return ok && proposer.Address == seq.Address
}

// SetNextProposer sets the next proposer for a rollapp
// called when the proposer has finished its notice period and rotation flow has started
func (k Keeper) setNextProposer(ctx sdk.Context, rollappId, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.NextProposerByRollappKey(rollappId)
	store.Set(nextProposerKey, addressBytes)
}

// GetNextProposer returns the next proposer for a rollapp
// It will return found=false if the next proposer is not set
// It will return found=true if the next proposer is set, even if it's empty
func (k Keeper) GetNextProposerLegacy(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.NextProposerByRollappKey(rollappId))
	if b == nil {
		return val, false
	}

	address := string(b)
	if address == types.SentinelSeqAddr {
		return val, true
	}
	return k.GetSequencer(ctx, address)
}

func (k Keeper) isNextProposerSet(ctx sdk.Context, rollappId string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.NextProposerByRollappKey(rollappId))
}

func (k Keeper) isNextProposer(ctx sdk.Context, seq types.Sequencer) bool {
	nextProposer, ok := k.GetNextProposer(ctx, seq.RollappId)
	return ok && nextProposer.Address == seq.Address
}

// removeNextProposer removes the next proposer for a rollapp
// called when the proposer has finished its rotation flow
func (k Keeper) removeNextProposer(ctx sdk.Context, rollappId string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.NextProposerByRollappKey(rollappId))
}
