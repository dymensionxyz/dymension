package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) time.Time {
	completionTime := ctx.BlockTime().Add(k.NoticePeriod(ctx))
	seq.NoticePeriodTime = completionTime

	k.UpdateSequencer(ctx, seq)
	k.AddSequencerToNoticePeriodQueue(ctx, seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.String()),
		),
	)

	return completionTime
}

// MatureSequencersWithNoticePeriod start rotation flow for all sequencers that have finished their notice period
// The next proposer is set to the next bonded sequencer
// The hub will expect a "last state update" from the sequencer to start unbonding
// In the middle of rotation, the next proposer required a notice period as well.
func (k Keeper) MatureSequencersWithNoticePeriod(ctx sdk.Context, currTime time.Time) {
	seqs := k.GetMatureNoticePeriodSequencers(ctx, currTime)
	for _, seq := range seqs {
		if k.isProposer(ctx, seq.RollappId, seq.Address) {
			k.startRotation(ctx, seq.RollappId)
			k.removeNoticePeriodSequencer(ctx, seq)
		}
		// next proposer cannot mature it's notice period until the current proposer has finished rotation
		// minor effect as notice_period >>> rotation time
	}
}

// IsRotating returns true if the rollapp is currently in the process of rotation.
// A process of rotation is defined by the existence of a next proposer. The next proposer can also be a "dummy" sequencer (i.e empty) in case no sequencer came. This is still considered rotation
// as the sequencer is rotating to an empty one (i.e gracefully leaving the rollapp).
// The next proposer can only be set after the notice period is over. The rotation period is over after the proposer sends his last batch.
func (k Keeper) IsRotating(ctx sdk.Context, rollappId string) bool {
	return k.isNextProposerSet(ctx, rollappId)
}

// IsProposerOrSuccessor returns true if the sequencer requires a notice period before unbonding
// Both the proposer and the next proposer require a notice period
func (k Keeper) IsProposerOrSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return k.isProposer(ctx, seq) || k.isNextProposer(ctx, seq)
}

// RequiresNoticePeriod returns true iff the sequencer requires a notice period before unbonding
func (k Keeper) RequiresNoticePeriod(ctx sdk.Context, seq types.Sequencer) bool {
	return k.IsProposerOrSuccessor(ctx, seq)
}

// startRotation sets the nextSequencer for the rollapp.
// This function will not clear the current proposer
// This function called when the sequencer has finished its notice period
func (k Keeper) startRotation(ctx sdk.Context, rollappId string) {
	// next proposer can be empty if there are no bonded sequencers available
	nextProposer := k.ExpectedNextProposer(ctx, rollappId)
	k.setNextProposer(ctx, rollappId, nextProposer.Address)

	k.Logger(ctx).Info("rotation started", "rollappId", rollappId, "nextProposer", nextProposer.Address)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotationStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeyNextProposer, nextProposer.Address),
		),
	)
}

// CompleteRotation completes the sequencer rotation flow.
// It's called when a last state update is received from the active, rotating sequencer.
// it will start unbonding the current proposer, and sets the nextProposer as the proposer.
func (k Keeper) CompleteRotation(ctx sdk.Context, rollappId string) error {
	proposer, ok := k.GetProposerLegacy(ctx, rollappId)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInternal, "proposer not set for rollapp %s", rollappId)
	}
	nextProposer, ok := k.GetNextProposer(ctx, rollappId)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInternal, "next proposer not set for rollapp %s", rollappId)
	}

	// start unbonding the current proposer
	k.startUnbondingPeriodForSequencer(ctx, &proposer)

	// change the proposer
	k.removeNextProposer(ctx, rollappId)
	k.SetProposer(ctx, rollappId, nextProposer.Address)

	if nextProposer.Address == types2.SentinelSeqAddr {
		k.Logger(ctx).Info("Rollapp left with no proposer.", "RollappID", rollappId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposer.Address),
		),
	)

	return nil
}

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

// SetSequencer set a specific sequencer in the store from its index
func (k Keeper) SetSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(
		sequencer.Address,
	), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, sequencer.Status)
	store.Set(seqByRollappKey, b)
}

// UpdateSequencer updates the state of a sequencer in the keeper.
// Parameters:
//   - sequencer: The sequencer object to be updated.
//   - oldStatus: An optional parameter representing the old status of the sequencer.
//     Needs to be provided if the status of the sequencer has changed (e.g from Bonded to Unbonding).
func (k Keeper) UpdateSequencer(ctx sdk.Context, sequencer *types.Sequencer, oldStatus ...types.OperatingStatus) {
	k.SetSequencer(ctx, *sequencer)

	// status changed, need to remove old status key
	if len(oldStatus) > 0 && sequencer.Status != oldStatus[0] {
		oldKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, oldStatus[0])
		ctx.KVStore(k.storeKey).Delete(oldKey)
	}
}

// GetSequencer returns a sequencer from its index
func (k Keeper) GetSequencer(ctx sdk.Context, sequencerAddress string) (val types.Sequencer, found bool) {
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

// MustGetSequencer returns a sequencer from its index
// It will panic if the sequencer is not found
func (k Keeper) MustGetSequencer(ctx sdk.Context, sequencerAddress string) types.Sequencer {
	seq, found := k.GetSequencer(ctx, sequencerAddress)
	if !found {
		panic("sequencer not found")
	}
	return seq
}

// GetAllSequencers returns all sequencer
func (k Keeper) GetAllSequencers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SequencersKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSequencersByRollapp returns a sequencersByRollapp from its index
func (k Keeper) GetSequencersByRollapp(ctx sdk.Context, rollappId string) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SequencersByRollappKey(rollappId))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSequencersByRollappByStatus returns a sequencersByRollapp from its index
func (k Keeper) GetSequencersByRollappByStatus(ctx sdk.Context, rollappId string, status types.OperatingStatus) (list []types.Sequencer) {
	prefixKey := types.SequencersByRollappByStatusKey(rollappId, status)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

/* -------------------------------------------------------------------------- */
/*                                notice period                               */
/* -------------------------------------------------------------------------- */

// GetMatureNoticePeriodSequencers returns all sequencers that have finished their notice period
func (k Keeper) GetMatureNoticePeriodSequencers(ctx sdk.Context, endTime time.Time) (list []types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.NoticePeriodQueueKey, sdk.PrefixEndBytes(types.NoticePeriodQueueByTimeKey(endTime)))

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// AddSequencerToNoticePeriodQueue set sequencer in notice period queue
func (k Keeper) AddSequencerToNoticePeriodQueue(ctx sdk.Context, sequencer *types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(sequencer)

	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.Address, sequencer.NoticePeriodTime)
	store.Set(noticePeriodKey, b)
}

// remove sequencer from notice period queue
func (k Keeper) removeNoticePeriodSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.Address, sequencer.NoticePeriodTime)
	store.Delete(noticePeriodKey)
}

/* ------------------------- proposer/next proposer ------------------------- */

// GetAllProposers returns all proposers for all rollapps
func (k Keeper) GetAllProposers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposerByRollappKey(""))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq := k.MustGetSequencer(ctx, address)
		list = append(list, seq)
	}

	return
}

func (k Keeper) SetProposer(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)

	activeKey := types.ProposerByRollappKey(rollapp)
	store.Set(activeKey, addressBytes)
}

func (k Keeper) GetProposer(ctx sdk.Context, rollappId string) (types.Sequencer, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ProposerByRollappKey(rollappId))
	if len(b) == 0 || string(b) == types2.SentinelSeqAddr {
		return types.Sequencer{
			Address:   types2.SentinelSeqAddr,
			RollappId: rollappId,
			// TODO: bonded status, other things?
		}, nil
	}
	s, ok := k.GetSequencer(ctx, string(b))
	if !ok {
		return types.Sequencer{}, gerrc.ErrInternal.Wrap("get sequencer")
	}
	return s, nil
}

// GetProposerLegacy returns the proposer for a rollapp
func (k Keeper) GetProposerLegacy(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ProposerByRollappKey(rollappId))
	if len(b) == 0 || string(b) == types2.SentinelSeqAddr {
		return val, false
	}

	return k.GetSequencer(ctx, string(b))
}

func (k Keeper) removeProposer(ctx sdk.Context, rollappId string) {
	k.SetProposer(ctx, rollappId, types2.SentinelSeqAddr)
}

func (k Keeper) isProposer(ctx sdk.Context, seq types.Sequencer) bool {
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
func (k Keeper) GetNextProposer(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.NextProposerByRollappKey(rollappId))
	if b == nil {
		return val, false
	}

	address := string(b)
	if address == types2.SentinelSeqAddr {
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
