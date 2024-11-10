package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// GetParams returns all of the parameters in the incentive module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.Params{
		MaxIterationsPerBlock: k.GetMaxIterationsPerBlock(ctx),
	}
}

// GetMaxIterationsPerBlock returns the maximum number of iterations per block.
func (k Keeper) GetMaxIterationsPerBlock(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, []byte(types.KeyMaxIterationsPerBlock), &res)
	return
}

// SetParams sets all of the parameters in the incentive module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
