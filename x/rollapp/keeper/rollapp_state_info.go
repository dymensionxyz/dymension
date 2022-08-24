package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// SetRollappStateInfo set a specific rollappStateInfo in the store from its index
func (k Keeper) SetRollappStateInfo(ctx sdk.Context, rollappStateInfo types.RollappStateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappStateInfoKeyPrefix))
	b := k.cdc.MustMarshal(&rollappStateInfo)
	store.Set(types.RollappStateInfoKey(
		rollappStateInfo.RollappId,
		rollappStateInfo.StateIndex,
	), b)
}

// GetRollappStateInfo returns a rollappStateInfo from its index
func (k Keeper) GetRollappStateInfo(
	ctx sdk.Context,
	rollappId string,
	stateIndex uint64,

) (val types.RollappStateInfo, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappStateInfoKeyPrefix))

	b := store.Get(types.RollappStateInfoKey(
		rollappId,
		stateIndex,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveRollappStateInfo removes a rollappStateInfo from the store
func (k Keeper) RemoveRollappStateInfo(
	ctx sdk.Context,
	rollappId string,
	stateIndex uint64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappStateInfoKeyPrefix))
	store.Delete(types.RollappStateInfoKey(
		rollappId,
		stateIndex,
	))
}

// GetAllRollappStateInfo returns all rollappStateInfo
func (k Keeper) GetAllRollappStateInfo(ctx sdk.Context) (list []types.RollappStateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappStateInfoKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.RollappStateInfo
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
