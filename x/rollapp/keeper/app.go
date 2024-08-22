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

func (k Keeper) GetApp(ctx sdk.Context, name, rollappId string) (val types.App, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(types.App{Name: name, RollappId: rollappId})
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

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		app := new(types.App)
		k.cdc.MustUnmarshal(iterator.Value(), app)
		list = append(list, app)
	}

	slices.SortFunc(list, func(a, b *types.App) int { return cmp.Compare(a.Order, b.Order) })

	return list
}
