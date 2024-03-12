package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// Called every block to finalize states that their dispute period over.
func (k Keeper) FinalizeQueue(ctx sdk.Context) {

	if uint64(ctx.BlockHeight()) < k.DisputePeriodInBlocks(ctx) {
		return
	}
	// check to see if there are pending  states to be finalized
	finalizationHeight := uint64(ctx.BlockHeight() - int64(k.DisputePeriodInBlocks(ctx)))
	pendingFinalizationQueue := k.GetAllFinalizationQueueUntilHeight(ctx, finalizationHeight)

	for _, blockHeightToFinalizationQueue := range pendingFinalizationQueue {

		// finalize pending states
		var failedToFinalizeQueue []types.StateInfoIndex
		for _, stateInfoIndex := range blockHeightToFinalizationQueue.FinalizationQueue {
			stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found || stateInfo.Status != common.Status_PENDING {
				ctx.Logger().Error("Missing stateInfo data when trying to finalize or alreay finalized", "rollappID", stateInfoIndex.RollappId, "height", ctx.BlockHeight(), "index", stateInfoIndex.Index)
				continue
			}

			wrappedFunc := func(ctx sdk.Context) error {
				stateInfo.Finalize()
				// update the status of the stateInfo
				k.SetStateInfo(ctx, stateInfo)
				// uppdate the LatestStateInfoIndex of the rollapp
				k.SetLatestFinalizedStateIndex(ctx, stateInfoIndex)
				// call the after-update-state hook
				keeperHooks := k.GetHooks()
				err := keeperHooks.AfterStateFinalized(ctx, stateInfoIndex.RollappId, &stateInfo)
				if err != nil {
					ctx.Logger().Error("Error after state finalized", "rollappID", stateInfoIndex.RollappId, "error", err.Error())
					return err
				}
				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeStateUpdate,
						stateInfo.GetEvents()...,
					),
				)
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				ctx.Logger().Error("Error finalizing state", "height", blockHeightToFinalizationQueue.CreationHeight, "rollappId", stateInfo.StateInfoIndex.RollappId)
				failedToFinalizeQueue = append(failedToFinalizeQueue, stateInfoIndex)
			}

		}
		k.RemoveBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue.CreationHeight)
		if len(failedToFinalizeQueue) > 0 {
			newBlockHeightToFinalizationQueue := types.BlockHeightToFinalizationQueue{
				CreationHeight:    blockHeightToFinalizationQueue.CreationHeight,
				FinalizationQueue: failedToFinalizeQueue}

			k.SetBlockHeightToFinalizationQueue(ctx, newBlockHeightToFinalizationQueue)
		}
	}
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

// GetAllFinalizationQueueUntilHeight returns all the blockHeightToFinalizationQueues with creation height equal or less to the input height
func (k Keeper) GetAllFinalizationQueueUntilHeight(ctx sdk.Context, height uint64) (list []types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.BlockHeightToFinalizationQueue
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.CreationHeight <= height {
			list = append(list, val)
		} else {
			break
		}
	}

	return
}

// GetAllBlockHeightToFinalizationQueue returns all blockHeightToFinalizationQueue
func (k Keeper) GetAllBlockHeightToFinalizationQueue(ctx sdk.Context) (list []types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.BlockHeightToFinalizationQueue
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
