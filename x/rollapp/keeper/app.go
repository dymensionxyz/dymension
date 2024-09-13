package keeper

import (
	"cmp"
	"slices"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) SetApp(ctx sdk.Context, app types.App) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(app)

	store.Set(key, k.cdc.MustMarshal(&app))
}

func (k Keeper) DeleteApp(ctx sdk.Context, app types.App) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(app)
	store.Delete(key)
}

func (k Keeper) GetApp(ctx sdk.Context, id uint64, rollappId string) (val types.App, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(types.App{Id: id, RollappId: rollappId})
	b := store.Get(key)
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetRollappApps(ctx sdk.Context, rollappId string) (list []*types.App) {
	var start []byte
	if rollappId != "" {
		start = types.RollappAppKeyPrefix(rollappId)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, start)

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		app := new(types.App)
		k.cdc.MustUnmarshal(iterator.Value(), app)
		list = append(list, app)
	}

	slices.SortFunc(list, func(a, b *types.App) int { return cmp.Compare(a.Order, b.Order) })

	return list
}

// GenerateNextAppID increments and returns the next available App ID for a specific Rollapp.
func (k Keeper) GenerateNextAppID(ctx sdk.Context, rollappID string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppSequenceKeyPrefix))
	sequenceKey := types.AppSequenceKey(rollappID)

	bz := store.Get(sequenceKey)
	var seq uint64
	if bz == nil {
		seq = 0
	} else {
		seq = sdk.BigEndianToUint64(bz)
	}

	seq++
	store.Set(sequenceKey, sdk.Uint64ToBigEndian(seq))

	return seq
}
