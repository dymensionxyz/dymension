package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// SetStateInfo set a specific stateInfo in the store from its index
func (k Keeper) SetStateInfo(ctx sdk.Context, stateInfo types.StateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	b := k.cdc.MustMarshal(&stateInfo)
	store.Set(types.StateInfoKey(
		stateInfo.StateInfoIndex,
	), b)

	// store a key prefixed with the creation timestamp
	storeTS := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TimestampedStateInfoKeyPrefix))
	storeTS.Set(types.StateInfoTimestampKey(
		stateInfo,
	), []byte{})
}

// GetStateInfo returns a stateInfo from its index
func (k Keeper) GetStateInfo(
	ctx sdk.Context,
	rollappId string,
	index uint64,
) (val types.StateInfo, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))

	b := store.Get(types.StateInfoKey(
		types.StateInfoIndex{RollappId: rollappId, Index: index},
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetLatestStateInfo is utility
func (k Keeper) GetLatestStateInfo(ctx sdk.Context,
	rollappId string,
) (types.StateInfo, bool) {
	ix, ok := k.GetLatestStateInfoIndex(ctx, rollappId)
	if !ok {
		return types.StateInfo{}, false
	}
	return k.GetStateInfo(ctx, rollappId, ix.GetIndex())
}

func (k Keeper) MustGetStateInfo(ctx sdk.Context,
	rollappId string,
	index uint64,
) (val types.StateInfo) {
	val, found := k.GetStateInfo(ctx, rollappId, index)
	if !found {
		panic(fmt.Sprintf("stateInfo not found for rollappId: %s, index: %d", rollappId, index))
	}
	return
}

// RemoveStateInfo removes a stateInfo from the store
func (k Keeper) RemoveStateInfo(
	ctx sdk.Context,
	rollappId string,
	index uint64,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	store.Delete(types.StateInfoKey(
		types.StateInfoIndex{RollappId: rollappId, Index: index},
	))
}

// GetAllStateInfo returns all stateInfo
func (k Keeper) GetAllStateInfo(ctx sdk.Context) (list []types.StateInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.StateInfo
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// DeleteStateInfoUntilTimestamp deletes all stateInfo until the given timestamp
func (k Keeper) DeleteStateInfoUntilTimestamp(ctx sdk.Context, endTimestampExcl time.Time) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StateInfoKeyPrefix))
	storeTS := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TimestampedStateInfoKeyPrefix))
	k.IterateStateInfoWithTimestamp(storeTS, endTimestampExcl.UnixMicro(), func(keyTS []byte) bool {
		key := keyTS[types.TimestampPrefixLen:] // remove the timestamp prefix
		store.Delete(key)
		storeTS.Delete(keyTS)
		return false
	})
}

// IterateStateInfoWithTimestamp iterates over stateInfo until timestamp
func (k Keeper) IterateStateInfoWithTimestamp(store prefix.Store, endTimestampUNIX int64, fn func(key []byte) (stop bool)) {
	endKey := types.StateInfoTimestampKeyPrefix(endTimestampUNIX)
	iterator := store.ReverseIterator(nil, endKey)

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		if fn(iterator.Key()) {
			break
		}
	}
}

// HasStateInfoTimestampKey checks if the stateInfo has a timestamp key
func (k Keeper) HasStateInfoTimestampKey(ctx sdk.Context, stateInfo types.StateInfo) bool {
	storeTS := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TimestampedStateInfoKeyPrefix))
	return storeTS.Has(types.StateInfoTimestampKey(stateInfo))
}
