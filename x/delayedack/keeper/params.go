package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// SetParams sets the module parameters in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, b)
}

// GetParams returns the module parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ParamsKey)
	if b == nil {
		panic("params should have been set")
	}

	k.cdc.MustUnmarshal(b, &params)
	return params
}

func (k Keeper) EpochIdentifier(ctx sdk.Context) (res string) {
	return k.GetParams(ctx).EpochIdentifier
}

func (k Keeper) BridgingFee(ctx sdk.Context) (res math.LegacyDec) {
	return k.GetParams(ctx).BridgingFee
}

func (k Keeper) BridgingFeeFromAmt(ctx sdk.Context, amt math.Int) (res math.Int) {
	return k.BridgingFee(ctx).MulInt(amt).TruncateInt()
}

func (k Keeper) DeletePacketsEpochLimit(ctx sdk.Context) (res int64) {
	return int64(k.GetParams(ctx).DeletePacketsEpochLimit)
}
