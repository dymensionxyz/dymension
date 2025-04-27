package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.KeyParams)

	var params types.Params
	k.cdc.MustUnmarshal(b, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&params)
	store.Set(types.KeyParams, b)
}

// DisputePeriodInBlocks returns the DisputePeriodInBlocks param
func (k Keeper) DisputePeriodInBlocks(ctx sdk.Context) (res uint64) {
	return k.GetParams(ctx).DisputePeriodInBlocks
}

func (k Keeper) LivenessSlashBlocks(ctx sdk.Context) (res uint64) {
	return k.GetParams(ctx).LivenessSlashBlocks
}

func (k Keeper) LivenessSlashInterval(ctx sdk.Context) (res uint64) {
	return k.GetParams(ctx).LivenessSlashInterval
}

// AppRegistrationFee returns the cost of adding an app
func (k Keeper) AppRegistrationFee(ctx sdk.Context) (res sdk.Coin) {
	return k.GetParams(ctx).AppRegistrationFee
}

func (k Keeper) MinSequencerBondGlobal(ctx sdk.Context) (res sdk.Coin) {
	return k.GetParams(ctx).MinSequencerBondGlobal
}
