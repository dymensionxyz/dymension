package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// LatestFinalizedStateIndex defines the rollapps' current (latest) index of the latest StateInfo that was finalized

// SetLatestFinalizedStateIndex set a specific latestFinalizedStateIndex in the store from its index
func (k Keeper) SetLatestFinalizedStateIndex(ctx sdk.Context, latestFinalizedStateIndex types.StateInfoIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateIndexKeyPrefix))
	b := k.cdc.MustMarshal(&latestFinalizedStateIndex)
	store.Set(types.LatestFinalizedStateIndexKey(
		latestFinalizedStateIndex.RollappId,
	), b)
}

// GetLatestFinalizedStateIndex returns a latestFinalizedStateIndex from its index
func (k Keeper) GetLatestFinalizedStateIndex(
	ctx sdk.Context,
	rollappId string,
) (val types.StateInfoIndex, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateIndexKeyPrefix))

	b := store.Get(types.LatestFinalizedStateIndexKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLatestFinalizedStateIndex removes a latestFinalizedStateIndex from the store
func (k Keeper) RemoveLatestFinalizedStateIndex(
	ctx sdk.Context,
	rollappId string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateIndexKeyPrefix))
	store.Delete(types.LatestFinalizedStateIndexKey(
		rollappId,
	))
}

// GetAllLatestFinalizedStateIndex returns latestFinalizedStateIndex for all rollapps
func (k Keeper) GetAllLatestFinalizedStateIndex(ctx sdk.Context) (list []types.StateInfoIndex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.LatestFinalizedStateIndexKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.StateInfoIndex
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetLatestFinalizedHeight(ctx sdk.Context, rollappID string) (uint64, error) {
	latestIndex, found := k.GetLatestFinalizedStateIndex(ctx, rollappID)
	if !found {
		return 0, gerrc.ErrNotFound.Wrapf("latest finalized state index is not found")
	}
	stateInfo := k.MustGetStateInfo(ctx, rollappID, latestIndex.Index)
	return stateInfo.GetLatestHeight(), nil
}

func (k Keeper) IsIndexFinalized(ctx sdk.Context, rollappID string, index uint64) bool {
	ixFinalized, ok := k.GetLatestFinalizedStateIndex(ctx, rollappID)
	if !ok {
		return false
	}
	return index <= ixFinalized.Index
}

func (k Keeper) IsHeightFinalized(ctx sdk.Context, rollappID string, height uint64) bool {
	latestFinalizedHeight, err := k.GetLatestFinalizedHeight(ctx, rollappID)
	if err != nil {
		return false
	}
	return height <= latestFinalizedHeight
}
