package keeper

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Called every block to finalize states that their dispute period over.
func (k Keeper) FinalizeQueue(ctx sdk.Context) {
	// check to see if there are pending  states to be finalized
	blockHeightToFinalizationQueue, found := k.GetBlockHeightToFinalizationQueue(ctx, uint64(ctx.BlockHeight()))
	if !found {
		return
	}

	// finalize pending states
	for _, stateInfoIndex := range blockHeightToFinalizationQueue.FinalizationQueue {
		stateInfo, found := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
		if !found {
			ctx.Logger().Error("Missing stateInfo data when trying to finalize", "rollappID", stateInfoIndex.RollappId, "height", ctx.BlockHeight(), "index", stateInfoIndex.Index)
			continue
		}
		stateInfo.Status = types.STATE_STATUS_FINALIZED
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
			sdk.NewEvent(types.EventTypeStatusChange,
				sdk.NewAttribute(types.AttributeKeyRollappId, stateInfoIndex.RollappId),
				sdk.NewAttribute(types.AttributeKeyStateInfoIndex, strconv.FormatUint(stateInfoIndex.Index, 10)),
				sdk.NewAttribute(types.AttributeKeyStartHeight, strconv.FormatUint(stateInfo.StartHeight, 10)),
				sdk.NewAttribute(types.AttributeKeyNumBlocks, strconv.FormatUint(stateInfo.NumBlocks, 10)),
				sdk.NewAttribute(types.AttributeKeyStatus, stateInfo.Status.String()),
			),
		)
	}
}

// SetBlockHeightToFinalizationQueue set a specific blockHeightToFinalizationQueue in the store from its index
func (k Keeper) SetBlockHeightToFinalizationQueue(ctx sdk.Context, blockHeightToFinalizationQueue types.BlockHeightToFinalizationQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	b := k.cdc.MustMarshal(&blockHeightToFinalizationQueue)
	store.Set(types.BlockHeightToFinalizationQueueKey(
		blockHeightToFinalizationQueue.FinalizationHeight,
	), b)
}

// GetBlockHeightToFinalizationQueue returns a blockHeightToFinalizationQueue from its index
func (k Keeper) GetBlockHeightToFinalizationQueue(
	ctx sdk.Context,
	finalizationHeight uint64,

) (val types.BlockHeightToFinalizationQueue, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))

	b := store.Get(types.BlockHeightToFinalizationQueueKey(
		finalizationHeight,
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
	finalizationHeight uint64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))
	store.Delete(types.BlockHeightToFinalizationQueueKey(
		finalizationHeight,
	))
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
