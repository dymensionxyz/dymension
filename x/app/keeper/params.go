package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.Cost(ctx),
	)
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// Cost returns the cost of creating an app
func (k Keeper) Cost(ctx sdk.Context) (res sdk.Coin) {
	k.paramstore.Get(ctx, types.KeyAppCost, &res)
	return
}
