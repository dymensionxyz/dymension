package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// MustGetParams returns the module parameters from the store
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
	if params.MovingAverageSmoothingFactor.IsNil() ||
		params.MovingAverageSmoothingFactor.IsNegative() ||
		params.MovingAverageSmoothingFactor.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("moving average smoothing factor must be between 0 and 1")
	}

	err := k.params.Set(ctx, params)
	if err != nil {
		return err
	}
	return nil
}
