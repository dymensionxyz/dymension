package keeper

import (
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// FinalizeRollappStates is called every block to finalize states when their dispute period over.
func (k Keeper) FinalizeRollappStates(ctx sdk.Context) {
	if uint64(ctx.BlockHeight()) < k.DisputePeriodInBlocks(ctx) {
		return
	}
	// check to see if there are pending  states to be finalized
	finalizationHeight := uint64(ctx.BlockHeight() - int64(k.DisputePeriodInBlocks(ctx)))
	queue, err := k.GetFinalizationQueueUntilHeightInclusive(ctx, finalizationHeight)
	if err != nil {
		panic(fmt.Errorf("get finalization queue until height: %d, %w", finalizationHeight, err))
	}

	k.FinalizeAllPending(ctx, queue)
}

func (k Keeper) FinalizeAllPending(ctx sdk.Context, pendingQueue []types.BlockHeightToFinalizationQueue) {
	// Cache the rollapps that failed to finalize at current EndBlocker execution.
	// The mapping is from rollappID to the first index of the state that failed to finalize.
	failedRollapps := make(map[string]uint64)
	// iterate over all the pending finalization height queues
	for _, queue := range pendingQueue {
		k.finalizeQueueForHeight(ctx, queue, failedRollapps)
	}
}

func (k Keeper) finalizeQueueForHeight(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue, failedRollapps map[string]uint64) {
	// finalize pending states
	for _, stateInfoIndex := range queue.FinalizationQueue {
		if _, failed := failedRollapps[stateInfoIndex.RollappId]; failed {
			// if the rollapp has already failed to finalize for at least one state index,
			// skip all subsequent state changes for this rollapp
			continue
		}
		if err := k.finalizeStateForIndex(ctx, stateInfoIndex); err != nil {
			// record the rollapp that had a failed state change at current height
			failedRollapps[stateInfoIndex.RollappId] = stateInfoIndex.Index
		}
	}
	k.updateQueueForHeight(ctx, queue, failedRollapps)
}

func (k Keeper) finalizeStateForIndex(ctx sdk.Context, stateInfoIndex types.StateInfoIndex) error {
	// if this fails, no state change will happen
	err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return k.finalizePending(ctx, stateInfoIndex)
	})
	if err != nil {
		// TODO: think about (non)recoverable errors and how to handle them accordingly
		k.Logger(ctx).Error(
			"failed to finalize state",
			"rollapp", stateInfoIndex.RollappId,
			"index", stateInfoIndex.Index,
			"error", err.Error(),
		)
		return fmt.Errorf("finalize state: %w", err)
	}
	return nil
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

func (k Keeper) updateQueueForHeight(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue, failedRollapps map[string]uint64) {
	// remove the queue if all the rollapp states are finalized
	if _, ok := failedRollapps[queue.RollappId]; !ok {
		k.RemoveFinalizationQueue(ctx, queue.CreationHeight, queue.RollappId)
		return
	}
	// remove from the queue only the indexes that were successfully finalized.
	// only delete the indexes up to the first failed state change.
	failedIndex := failedRollapps[queue.RollappId]
	queue.FinalizationQueue = slices.DeleteFunc(
		queue.FinalizationQueue,
		func(si types.StateInfoIndex) bool {
			return si.Index < failedIndex
		},
	)
	// save the current queue with "leftover" rollapp state changes
	k.SetFinalizationQueue(ctx, queue)
}

// SetFinalizationQueue set types.BlockHeightToFinalizationQueue for a specific height and rollappID.
// Panics on encoding errors.
func (k Keeper) SetFinalizationQueue(ctx sdk.Context, queue types.BlockHeightToFinalizationQueue) {
	err := k.finalizationQueue.Set(ctx, collections.Join(queue.CreationHeight, queue.RollappId), queue)
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
// Panics on encoding errors. Do not panic if the key does not exist.
func (k Keeper) RemoveFinalizationQueue(ctx sdk.Context, height uint64, rollappID string) {
	err := k.finalizationQueue.Remove(ctx, collections.Join(height, rollappID))
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

// GetAllFinalizationQueueUntilHeightInclusive returns all the blockHeightToFinalizationQueues with creation height equal or less to the input height
// Deprecated: use GetFinalizationQueueUntilHeightInclusive instead. Only used in state migrations.
func (k Keeper) GetAllFinalizationQueueUntilHeightInclusive(ctx sdk.Context, height uint64) (list []types.BlockHeightToFinalizationQueue) {
	height++
	return k.getFinalizationQueue(ctx, &height)
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
