package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// LatestStateInfoIndex defines the rollapps' current (latest) index of the last UpdateState

// SetLatestStateInfoIndex set a specific latestStateInfoIndex in the store from its index
func (k Keeper) SetLatestStateInfoIndex(ctx sdk.Context, latestStateInfoIndex types.StateInfoIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestStateInfoIndexKeyPrefix))
	b := k.cdc.MustMarshal(&latestStateInfoIndex)
	store.Set(types.LatestStateInfoIndexKey(
		latestStateInfoIndex.RollappId,
	), b)
}

// GetLatestStateInfoIndex returns a latestStateInfoIndex from its index
func (k Keeper) GetLatestStateInfoIndex(
	ctx sdk.Context,
	rollappId string,
) (val types.StateInfoIndex, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestStateInfoIndexKeyPrefix))

	b := store.Get(types.LatestStateInfoIndexKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLatestStateInfoIndex removes a latestStateInfoIndex from the store
func (k Keeper) RemoveLatestStateInfoIndex(
	ctx sdk.Context,
	rollappId string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestStateInfoIndexKeyPrefix))
	store.Delete(types.LatestStateInfoIndexKey(
		rollappId,
	))
}

// GetAllLatestStateInfoIndex returns all latestStateInfoIndex
func (k Keeper) GetAllLatestStateInfoIndex(ctx sdk.Context) (list []types.StateInfoIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestStateInfoIndexKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.StateInfoIndex
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
