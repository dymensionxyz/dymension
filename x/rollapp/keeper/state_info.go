package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// SetStateInfo set a specific stateInfo in the store from its index
func (k Keeper) SetStateInfo(ctx sdk.Context, stateInfo types.StateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	b := k.cdc.MustMarshal(&stateInfo)
	store.Set(types.StateInfoKey(
		stateInfo.RollappId,
		stateInfo.StateIndex,
	), b)
}

// GetStateInfo returns a stateInfo from its index
func (k Keeper) GetStateInfo(
	ctx sdk.Context,
	rollappId string,
	stateIndex uint64,

) (val types.StateInfo, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))

	b := store.Get(types.StateInfoKey(
		rollappId,
		stateIndex,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveStateInfo removes a stateInfo from the store
func (k Keeper) RemoveStateInfo(
	ctx sdk.Context,
	rollappId string,
	stateIndex uint64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	store.Delete(types.StateInfoKey(
		rollappId,
		stateIndex,
	))
}

// GetAllStateInfo returns all stateInfo
func (k Keeper) GetAllStateInfo(ctx sdk.Context) (list []types.StateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	// nolint: errcheck
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StateInfo
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
