package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.EpochIdentifier(ctx))
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) EpochIdentifier(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyEpochIdentifier, &res)
	return
}

func (k Keeper) TimeoutFee(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyTimeoutFee, &res)
	return
}

func (k Keeper) ErrAckFee(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyErrAckFee, &res)
	return
}
