package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) BridgingFee(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyBridgeFee, &res)
	return
}

func (k Keeper) BridgingFeeFromAmt(ctx sdk.Context, amt sdk.Int) (res sdk.Int) {
	return k.BridgingFee(ctx).MulInt(amt).TruncateInt()
}

func (k Keeper) DeletePacketsEpochLimit(ctx sdk.Context) (res int64) {
	k.paramstore.Get(ctx, types.KeyDeletePacketsEpochLimit, &res)
	return
}
