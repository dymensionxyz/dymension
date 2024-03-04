package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Called every block to finalize states that their dispute period over.
func (k Keeper) FinalizeQueue(ctx sdk.Context) {
	// check to see if there are pending  states to be finalized
	pendingFinalizationQueue := k.GetPendingFinalizationQueue(ctx, uint64(ctx.BlockHeight()-int64(k.DisputePeriodInBlocks(ctx))))

	for _, blockHeightToFinalizationQueue := range pendingFinalizationQueue {

		// finalize pending states
		for _, stateInfoIndex := range blockHeightToFinalizationQueue.FinalizationQueue {
			stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found {
				ctx.Logger().Error("Missing stateInfo data when trying to finalize", "rollappID", stateInfoIndex.RollappId, "height", ctx.BlockHeight(), "index", stateInfoIndex.Index)
				continue
			}
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
			}

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeStateUpdate,
					stateInfo.GetEvents()...,
				),
			)

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

// GetPendingFinalizationQueue returns the blockHeightToFinalizationQueues starting from the input height (height after the disputeperiod) till the last queue with batches not yet finalized
func (k Keeper) GetPendingFinalizationQueue(ctx sdk.Context, height uint64) (list []types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	heightKey := types.BlockHeightToFinalizationQueueKey(height + 1)
	iterator := sdk.KVStoreReversePrefixIterator(store, heightKey)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.BlockHeightToFinalizationQueue
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		stateInfoIndex := val.FinalizationQueue
		stateInfo, _ := k.GetStateInfo(ctx, stateInfoIndex[0].RollappId, stateInfoIndex[0].Index)
		if stateInfo.Status == types.STATE_STATUS_FINALIZED {
			break
		}
		list = append(list, val)
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
