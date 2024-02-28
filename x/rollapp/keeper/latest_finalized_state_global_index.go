package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// LatestFinalizedStateIndex defines the rollapps' current (latest) index of the latest StateInfo that was finalized

// SetLatestFinalizedStateIndex set a specific latestFinalizedStateIndex in the store from its index
func (k Keeper) SetLatestFinalizedGlobalStateIndex(ctx sdk.Context, latestFinalizedStateGlobalIndex types.StateInfoGlobalIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateGlobalIndexKeyPrefix))
	b := k.cdc.MustMarshal(&latestFinalizedStateGlobalIndex)
	store.Set(types.LatestFinalizedStateGlobalIndexKey(), b)
}

// GetLatestFinalizedStateIndex returns a latestFinalizedStateIndex from its index
func (k Keeper) GetLatestFinalizedStateGlobalIndex(ctx sdk.Context) (val types.StateInfoGlobalIndex, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateGlobalIndexKeyPrefix))

	b := store.Get(types.LatestFinalizedStateGlobalIndexKey())
	if b == nil {
		return types.StateInfoGlobalIndex{Index: 0}, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLatestFinalizedStateIndex removes a latestFinalizedStateIndex from the store
func (k Keeper) RemoveLatestFinalizedStateGlobalIndex(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateGlobalIndexKeyPrefix))
	store.Delete(types.LatestFinalizedStateGlobalIndexKey())
}
