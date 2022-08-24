package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// SetStateIndex set a specific stateIndex in the store from its index
func (k Keeper) SetStateIndex(ctx sdk.Context, stateIndex types.StateIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateIndexKeyPrefix))
	b := k.cdc.MustMarshal(&stateIndex)
	store.Set(types.StateIndexKey(
		stateIndex.RollappId,
	), b)
}

// GetStateIndex returns a stateIndex from its index
func (k Keeper) GetStateIndex(
	ctx sdk.Context,
	rollappId string,

) (val types.StateIndex, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateIndexKeyPrefix))

	b := store.Get(types.StateIndexKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveStateIndex removes a stateIndex from the store
func (k Keeper) RemoveStateIndex(
	ctx sdk.Context,
	rollappId string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateIndexKeyPrefix))
	store.Delete(types.StateIndexKey(
		rollappId,
	))
}

// GetAllStateIndex returns all stateIndex
func (k Keeper) GetAllStateIndex(ctx sdk.Context) (list []types.StateIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateIndexKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StateIndex
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
