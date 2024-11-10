package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns all the parameters in the incentive module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.Params{
		DistrEpochIdentifier: k.DistrEpochIdentifier(ctx),
		CreateGaugeBaseFee:   k.CreateGaugeBaseFee(ctx),
		AddToGaugeBaseFee:    k.AddToGaugeBaseFee(ctx),
		AddDenomFee:          k.AddDenomFee(ctx),
	}
}

func (k Keeper) DistrEpochIdentifier(ctx sdk.Context) (res string) {
	k.paramSpace.Get(ctx, types.KeyDistrEpochIdentifier, &res)
	return
}

func (k Keeper) CreateGaugeBaseFee(ctx sdk.Context) (res sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyCreateGaugeFee, &res)
	return
}

func (k Keeper) AddToGaugeBaseFee(ctx sdk.Context) (res sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyAddToGaugeFee, &res)
	return
}

func (k Keeper) AddDenomFee(ctx sdk.Context) (res sdk.Int) {
	k.paramSpace.Get(ctx, types.KeyAddDenomFee, &res)
	return
}

// SetParams sets all of the parameters in the incentive module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
