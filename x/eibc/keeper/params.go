package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ParamsKey)

	var params types.Params
	k.cdc.MustUnmarshal(b, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, b)
}

func (k Keeper) EpochIdentifier(ctx sdk.Context) (res string) {
	return k.GetParams(ctx).EpochIdentifier
}

func (k Keeper) TimeoutFee(ctx sdk.Context) (res math.LegacyDec) {
	return k.GetParams(ctx).TimeoutFee
}

func (k Keeper) ErrAckFee(ctx sdk.Context) (res math.LegacyDec) {
	return k.GetParams(ctx).ErrackFee
}
