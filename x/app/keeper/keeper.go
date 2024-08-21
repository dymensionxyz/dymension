package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramstore paramtypes.Subspace

	rollappKeeper types.RollappKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	rollappKeeper types.RollappKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramstore:    ps,
		rollappKeeper: rollappKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetApp(ctx sdk.Context, app types.App) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(app.Name, app.RollappId)
	store.Set(key, k.cdc.MustMarshal(&app))
}

func (k Keeper) RemoveApp(ctx sdk.Context, app types.App) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(app.Name, app.RollappId)
	store.Delete(key)
}

func (k Keeper) GetApp(ctx sdk.Context, name, rollappId string) (val types.App, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	key := types.AppKey(name, rollappId)
	b := store.Get(key)
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetAllApps(ctx sdk.Context, rollappId string) (list []types.App) {
	var start []byte
	if rollappId != "" {
		start = []byte(rollappId)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AppKeyPrefix))
	iterator := store.Iterator(start, nil)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var app types.App
		k.cdc.MustUnmarshal(iterator.Value(), &app)
		list = append(list, app)
	}
	return list
}
