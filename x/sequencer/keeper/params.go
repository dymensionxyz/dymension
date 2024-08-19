package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MinBond(ctx),
		k.UnbondingTime(ctx),
		k.NoticePeriod(ctx),
		k.LivenessSlashMultiplier(ctx),
	)
}

func (k Keeper) MinBond(ctx sdk.Context) (res sdk.Coin) {
	k.paramstore.Get(ctx, types.KeyMinBond, &res)
	return
}

func (k Keeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyUnbondingTime, &res)
	return
}

func (k Keeper) NoticePeriod(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyNoticePeriod, &res)
	return
}

func (k Keeper) LivenessSlashMultiplier(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyLivenessSlashMultiplier, &res)
	return
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
