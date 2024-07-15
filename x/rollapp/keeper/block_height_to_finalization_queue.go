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

	k.FinalizePending(ctx, pendingFinalizationQueue)
}

func (k Keeper) FinalizePending(ctx sdk.Context, pendingFinalizationQueue []types.BlockHeightToFinalizationQueue) {
	// Iterate over all the pending finalization queue
	for _, blockHeightToFinalizationQueue := range pendingFinalizationQueue {
		failedRollapps := make(map[string][]uint64)
		// finalize pending states
		for _, stateInfoIndex := range blockHeightToFinalizationQueue.FinalizationQueue {
			if _, failed := failedRollapps[stateInfoIndex.RollappId]; failed {
				// if the rollapp has already failed to finalize at current height for at least one state index,
				// add all subsequent indices and move on
				failedRollapps[stateInfoIndex.RollappId] = append(failedRollapps[stateInfoIndex.RollappId], stateInfoIndex.Index)
				continue
			}
			// if this fails, no state change will happen
			err := osmoutils.ApplyFuncIfNoError(ctx,
				func(ctx sdk.Context) error {
					return k.finalizePending(ctx, stateInfoIndex)
				})
			if err != nil {
				// record the rollapp that had a failed state change at current height
				failedRollapps[stateInfoIndex.RollappId] = append(failedRollapps[stateInfoIndex.RollappId], stateInfoIndex.Index)
				// TODO: think about (non)recoverable errors and how to handle them accordingly
				k.Logger(ctx).Error("failed to finalize state", "rollapp", stateInfoIndex.RollappId, "index", stateInfoIndex.Index, "error", err.Error())
				continue
			}
		}
		// remove the blockHeightToFinalizationQueue if all the rollapps' states are finalized
		if len(failedRollapps) == 0 {
			k.RemoveBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue.CreationHeight)
		} else {
			// remove from the queue only the rollapps that were successfully finalized at all indices.
			// for failed rollapps at current height, all state indices including and following the failed index will stay in the queue
			blockHeightToFinalizationQueue.FinalizationQueue = slices.DeleteFunc(blockHeightToFinalizationQueue.FinalizationQueue,
				func(si types.StateInfoIndex) bool {
					idxs, ok := failedRollapps[si.RollappId]
					return !(ok && slices.Contains(idxs, si.Index))
				})
			// save the current queue with "leftover" rollapp's state changes
			k.SetBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue)
		}
	}
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
