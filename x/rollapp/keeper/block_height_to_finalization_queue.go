package keeper

import (
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) CanUnbond(ctx sdk.Context, seq sequencertypes.Sequencer) error {
	rng := collections.NewPrefixedPairRange[string, uint64](seq.Address)
	return k.seqToUnfinalizedHeight.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (stop bool, err error) {
		// we found one!
		return true, errorsmod.Wrapf(sequencertypes.ErrUnbondNotAllowed, "unfinalized height: h: %d", key.K2())
	})
}

// PruneSequencerHeights removes bookkeeping for all heights ABOVE h for given sequencers
// On rollback, this should be called passing all sequencers who sequenced a rolled back block
func (k Keeper) PruneSequencerHeights(ctx sdk.Context, sequencers []string, h uint64) error {
	for _, seqAddr := range sequencers {
		rng := collections.NewPrefixedPairRange[string, uint64](seqAddr).StartExclusive(h)
		if err := k.seqToUnfinalizedHeight.Clear(ctx, rng); err != nil {
			return errorsmod.Wrapf(err, "seq: %s", seqAddr)
		}
	}
	return nil
}

func (k Keeper) SaveSequencerHeight(ctx sdk.Context, seqAddr string, height uint64) error {
	return k.seqToUnfinalizedHeight.Set(ctx, collections.Join(seqAddr, height))
}

func (k Keeper) DelSequencerHeight(ctx sdk.Context, seqAddr string, height uint64) error {
	return k.seqToUnfinalizedHeight.Remove(ctx, collections.Join(seqAddr, height))
}

func (k Keeper) AllSequencerHeightPairs(ctx sdk.Context) ([]types.SequencerHeightPair, error) {
	ret := make([]types.SequencerHeightPair, 0)
	err := k.seqToUnfinalizedHeight.Walk(ctx, nil, func(key collections.Pair[string, uint64]) (stop bool, err error) {
		ret = append(ret, types.SequencerHeightPair{Sequencer: key.K1(), Height: key.K2()})
		return false, nil
	})
	return ret, err
}

// FinalizeRollappStates is called every block to finalize states when their dispute period over.
func (k Keeper) FinalizeRollappStates(ctx sdk.Context) {
	if uint64(ctx.BlockHeight()) < k.DisputePeriodInBlocks(ctx) {
		return
	}
	// check to see if there are pending  states to be finalized
	finalizationHeight := uint64(ctx.BlockHeight() - int64(k.DisputePeriodInBlocks(ctx)))
	queue, err := k.GetFinalizationQueueUntilHeightInclusive(ctx, finalizationHeight)
	if err != nil {
		// The error is returned only if there is an internal issue with the store iterator or encoding.
		// This should never happen in practice.
		k.Logger(ctx).With("error", err, "height", finalizationHeight).
			Error("failed to get finalization queue until height")
		return
	}

	k.FinalizeAllPending(ctx, queue)
}

// FinalizeAllPending is called every block to finalize all pending states in the queue.
// pendingQueues contains queues in ascending order of creation height. There may be multiple queues for the same
// creation height since multiple rollapps may have pending states. In that case, the queues are ordered by rollappID.
// If one of rollapps states fails to finalize, the rest of the states are not finalized as well. This is achieved by
// using a set of failed rollapps.
func (k Keeper) FinalizeAllPending(ctx sdk.Context, pendingQueues []types.BlockHeightToFinalizationQueue) {
	// Cache the rollapps that failed to finalize at current EndBlocker execution.
	// Once the rollapp is added to this set, the rest of the states for that rollapp are not finalized.
	// Map here is safe to use since it's used only as a key set for existence checks.
	failedRollapps := make(map[string]struct{})

	// Iterate over all the pending finalization height queues
	for _, queue := range pendingQueues {
		// Check if the rollapp failed to finalize previously
		_, failed := failedRollapps[queue.RollappId]
		if failed {
			// Skip the rollapp if it failed to finalize previously
			continue
		}

		// Finalize the queue. If it fails, add the rollapp to the failedRollapps set.
		finalized := k.FinalizeStates(ctx, queue)
		if !finalized {
			failedRollapps[queue.RollappId] = struct{}{}
		}
	}
}

// FinalizeStates finalizes all the pending states in the queue. Returns true if all the states are finalized successfully.
func (k Keeper) FinalizeStates(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue) bool {
	for i, stateInfoIndex := range queue.FinalizationQueue {
		// if this fails, no state change will happen
		err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.finalizePending(ctx, stateInfoIndex)
		})
		if err != nil {
			k.Logger(ctx).
				With("rollapp_id", stateInfoIndex.RollappId, "index", stateInfoIndex.Index, "err", err.Error()).
				Error("failed to finalize rollapp state")
			// remove from the queue only the indexes that were successfully finalized.
			// delete up to the first failed state change, exclusively.
			queue.FinalizationQueue = slices.Delete(queue.FinalizationQueue, 0, i)

			// Save the current queue with "leftover" rollapp state changes.
			//
			// Use Must-method here as there are no any other options to handle the error.
			// At this stage, some of the states are already finalized. If we can't save the queue, then
			// the runtime will panic on the next EndBlocker when we try to finalize the same states from
			// the queue again. Plus, if the panic occurs here, it means that there is an internal issue
			// with the store. This should never happen in practice.
			k.MustSetFinalizationQueue(ctx, queue)

			return false
		}
	}

	// Remove the queue if all the states are finalized.
	//
	// Use Must-method here as there are no any other options to handle the error.
	// At this stage, all the states are already finalized. If we can't remove the queue, then
	// the runtime will panic on the next EndBlocker when we try to finalize the same states from
	// the queue again. Plus, if the panic occurs here, it means that there is an internal issue
	// with the store. This should never happen in practice.
	k.MustRemoveFinalizationQueue(ctx, queue.CreationHeight, queue.RollappId)

	return true
}

func (k *Keeper) finalizePendingState(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error {
	stateInfo := k.MustGetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
	if stateInfo.Status != common.Status_PENDING {
		panic(fmt.Sprintf("invariant broken: stateInfo is not in pending state: rollapp: %s: status: %s", stateInfoIndex.RollappId, stateInfo.Status))
	}
	stateInfo.Finalize()
	// update the status of the stateInfo
	k.SetStateInfo(ctx, stateInfo)
	// update the LatestStateInfoIndex of the rollapp
	k.SetLatestFinalizedStateIndex(ctx, stateInfoIndex)

	for _, bd := range stateInfo.BDs.BD {
		if err := k.DelSequencerHeight(ctx, stateInfo.Sequencer, bd.Height); err != nil {
			return errorsmod.Wrap(err, "del sequencer height")
		}
	}

	// call the after-update-state hook
	err := k.GetHooks().AfterStateFinalized(ctx, stateInfoIndex.RollappId, &stateInfo)
	if err != nil {
		return fmt.Errorf("after state finalized: %w", err)
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStatusChange,
			stateInfo.GetEvents()...,
		),
	)
	return nil
}

// SetFinalizationQueue set types.BlockHeightToFinalizationQueue for a specific height and rollappID.
func (k Keeper) SetFinalizationQueue(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue) error {
	return k.finalizationQueue.Set(ctx, collections.Join(queue.CreationHeight, queue.RollappId), queue)
}

// MustSetFinalizationQueue is a wrapper for SetFinalizationQueue that panics on error.
// Panics only on encoding errors (this implies from the internal implementation).
func (k Keeper) MustSetFinalizationQueue(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue) {
	err := k.SetFinalizationQueue(ctx, queue)
	if err != nil {
		panic(err)
	}
}

// GetFinalizationQueue gets types.BlockHeightToFinalizationQueue for a specific height and rollappID.
// Panics on encoding errors.
func (k Keeper) GetFinalizationQueue(ctx sdk.Context, height uint64, rollappID string) (types.BlockHeightToFinalizationQueue, bool) {
	queue, err := k.finalizationQueue.Get(ctx, collections.Join(height, rollappID))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	found := err == nil
	return queue, found
}

// RemoveFinalizationQueue removes types.BlockHeightToFinalizationQueue.
func (k Keeper) RemoveFinalizationQueue(ctx sdk.Context, height uint64, rollappID string) error {
	return k.finalizationQueue.Remove(ctx, collections.Join(height, rollappID))
}

// MustRemoveFinalizationQueue is a wrapper for RemoveFinalizationQueue that panics on error.
// Panics on encoding errors. Do not panic if the key does not exist (this implies from the internal implementation).
func (k Keeper) MustRemoveFinalizationQueue(ctx sdk.Context, height uint64, rollappID string) {
	err := k.RemoveFinalizationQueue(ctx, height, rollappID)
	if err != nil {
		panic(err)
	}
}

// GetFinalizationQueueUntilHeightInclusive returns all types.BlockHeightToFinalizationQueue with creation height equal or less to the input height
func (k Keeper) GetFinalizationQueueUntilHeightInclusive(ctx sdk.Context, height uint64) ([]types.BlockHeightToFinalizationQueue, error) {
	rng := collections.NewPrefixUntilPairRange[uint64, string](height)
	iter, err := k.finalizationQueue.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close() // nolint: errcheck
	return iter.Values()
}

// GetFinalizationQueueByRollapp returns all states from different heights associated with a given rollapp
func (k Keeper) GetFinalizationQueueByRollapp(ctx sdk.Context, rollapp string) ([]types.BlockHeightToFinalizationQueue, error) {
	iter, err := k.finalizationQueue.Indexes.RollappIDReverseLookup.MatchExact(ctx, rollapp)
	if err != nil {
		return nil, err
	}
	defer iter.Close() // nolint: errcheck
	var res []types.BlockHeightToFinalizationQueue
	for ; iter.Valid(); iter.Next() {
		key, err := iter.PrimaryKey()
		if err != nil {
			return nil, err
		}
		queue, err := k.finalizationQueue.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		res = append(res, queue)
	}
	return res, nil
}

// GetEntireFinalizationQueue returns all types.BlockHeightToFinalizationQueue
func (k Keeper) GetEntireFinalizationQueue(ctx sdk.Context) ([]types.BlockHeightToFinalizationQueue, error) {
	iter, err := k.finalizationQueue.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close() // nolint: errcheck
	return iter.Values()
}

// SetBlockHeightToFinalizationQueue set a specific blockHeightToFinalizationQueue in the store from its index
// Deprecated: use SetFinalizationQueue instead. Only used in state migrations.
func (k Keeper) SetBlockHeightToFinalizationQueue(ctx sdk.Context, blockHeightToFinalizationQueue types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	b := k.cdc.MustMarshal(&blockHeightToFinalizationQueue)
	store.Set(types.BlockHeightToFinalizationQueueKey(
		blockHeightToFinalizationQueue.CreationHeight,
	), b)
}

// RemoveBlockHeightToFinalizationQueue removes a blockHeightToFinalizationQueue from the store
// Deprecated: use RemoveFinalizationQueue instead. Only used in state migrations.
func (k Keeper) RemoveBlockHeightToFinalizationQueue(
	ctx sdk.Context,
	creationHeight uint64,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	store.Delete(types.BlockHeightToFinalizationQueueKey(
		creationHeight,
	))
}

// GetAllBlockHeightToFinalizationQueue returns all blockHeightToFinalizationQueue
// Deprecated: use GetEntireFinalizationQueue instead. Only used in state migrations.
func (k Keeper) GetAllBlockHeightToFinalizationQueue(ctx sdk.Context) (list []types.BlockHeightToFinalizationQueue) {
	return k.getFinalizationQueue(ctx, nil)
}

// Deprecated: only used in GetAllFinalizationQueueUntilHeightInclusive and GetAllBlockHeightToFinalizationQueue
func (k Keeper) getFinalizationQueue(ctx sdk.Context, endHeightNonInclusive *uint64) (list []types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.BlockHeightToFinalizationQueue
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if endHeightNonInclusive != nil && *endHeightNonInclusive <= val.CreationHeight {
			break
		}
		list = append(list, val)
	}

	return
}
