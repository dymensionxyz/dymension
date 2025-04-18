package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns all of the parameters in the incentive module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams sets all of the parameters in the incentive module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) DistrEpochIdentifier(ctx sdk.Context) (res string) {
	k.paramSpace.Get(ctx, types.KeyDistrEpochIdentifier, &res)
	return
}

func (k Keeper) RollappGaugesMode(ctx sdk.Context) (res types.Params_RollappGaugesModes) {
	k.paramSpace.Get(ctx, types.KeyRollappGaugesMode, &res)
	return
}
