package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/auctionhouse/types"
)

func (k Keeper) MustGetParams(ctx sdk.Context) types.Params {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return params
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.ValidateBasic(); err != nil {
		return err
	}

	// FIXME: validate accepted tokens have txfees route to basedenom

	return k.params.Set(ctx, params)
}
