package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.DisputePeriodInBlocks(ctx),
		k.LivenessSlashBlocks(ctx),
		k.LivenessSlashInterval(ctx),
		k.AppRegistrationFee(ctx),
		k.MinSequencerBondGlobal(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// DisputePeriodInBlocks returns the DisputePeriodInBlocks param
func (k Keeper) DisputePeriodInBlocks(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyDisputePeriodInBlocks, &res)
	return
}

func (k Keeper) LivenessSlashBlocks(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyLivenessSlashBlocks, &res)
	return
}

func (k Keeper) LivenessSlashInterval(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyLivenessSlashInterval, &res)
	return
}

// AppRegistrationFee returns the cost of adding an app
func (k Keeper) AppRegistrationFee(ctx sdk.Context) (res sdk.Coin) {
	k.paramstore.Get(ctx, types.KeyAppRegistrationFee, &res)
	return
}

func (k Keeper) MinSequencerBondGlobal(ctx sdk.Context) (res sdk.Coin) {
	k.paramstore.Get(ctx, types.KeyMinSequencerBondGlobal, &res)
	return
}
