package keeper

import (
	"fmt"
	"slices"

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
	pendingFinalizationQueue := k.GetAllFinalizationQueueUntilHeightInclusive(ctx, finalizationHeight)

	k.FinalizeAllPending(ctx, pendingFinalizationQueue)
}

func (k Keeper) FinalizeAllPending(ctx sdk.Context, pendingFinalizationQueue []types.BlockHeightToFinalizationQueue) {
	// Cache the rollapps that failed to finalize at current EndBlocker execution.
	// The mapping is from rollappID to the first index of the state that failed to finalize.
	failedRollapps := make(map[string]uint64)
	// iterate over all the pending finalization height queues
	for _, blockHeightToFinalizationQueue := range pendingFinalizationQueue {
		k.finalizeQueueForHeight(ctx, blockHeightToFinalizationQueue, failedRollapps)
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
	keeperHooks := k.GetHooks()
	err := keeperHooks.AfterStateFinalized(ctx, stateInfoIndex.RollappId, &stateInfo)
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

func (k Keeper) updateQueueForHeight(ctx sdk.Context, blockHeightToFinalizationQueue types.BlockHeightToFinalizationQueue, failedRollapps map[string]uint64) {
	// remove the blockHeightToFinalizationQueue if all the rollapps' states are finalized
	if len(failedRollapps) == 0 {
		k.RemoveBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue.CreationHeight)
		return
	}
	// remove from the queue only the rollapps that were successfully finalized at all indices.
	// while iterating the queue for deleting the successfully finalized states, we remove them if
	// - rollapp was not found in the failedRollapps map
	// - if it was found, the indexes to be deleted should only be the ones up to the point of the index of the first failed state change
	blockHeightToFinalizationQueue.FinalizationQueue = slices.DeleteFunc(blockHeightToFinalizationQueue.FinalizationQueue,
		func(si types.StateInfoIndex) bool {
			idx, failed := failedRollapps[si.RollappId]
			return !failed || si.Index < idx
		})
	// save the current queue with "leftover" rollapp's state changes
	k.SetBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue)
}

// SetBlockHeightToFinalizationQueue set a specific blockHeightToFinalizationQueue in the store from its index
func (k Keeper) SetBlockHeightToFinalizationQueue(ctx sdk.Context, blockHeightToFinalizationQueue types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	b := k.cdc.MustMarshal(&blockHeightToFinalizationQueue)
	store.Set(types.BlockHeightToFinalizationQueueKey(
		blockHeightToFinalizationQueue.CreationHeight,
	), b)
}

// GetBlockHeightToFinalizationQueue returns a blockHeightToFinalizationQueue from its index
func (k Keeper) GetBlockHeightToFinalizationQueue(
	ctx sdk.Context,
	creationHeight uint64,
) (val types.BlockHeightToFinalizationQueue, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))

	b := store.Get(types.BlockHeightToFinalizationQueueKey(
		creationHeight,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlockHeightToFinalizationQueue removes a blockHeightToFinalizationQueue from the store
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
func (k Keeper) GetAllFinalizationQueueUntilHeightInclusive(ctx sdk.Context, height uint64) (list []types.BlockHeightToFinalizationQueue) {
	height++
	return k.getFinalizationQueue(ctx, &height)
}

// GetAllBlockHeightToFinalizationQueue returns all blockHeightToFinalizationQueue
func (k Keeper) GetAllBlockHeightToFinalizationQueue(ctx sdk.Context) (list []types.BlockHeightToFinalizationQueue) {
	return k.getFinalizationQueue(ctx, nil)
}

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
