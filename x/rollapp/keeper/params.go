package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.RollappsEnabled(ctx),
		k.DisputePeriodInBlocks(ctx),
		k.DeployerWhitelist(ctx),
		k.HubExpectedBlockTime(ctx),
		k.LivenessSlashTime(ctx),
		k.LivenessSlashInterval(ctx),
		k.LivenessJailTime(ctx),
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

// DeployerWhitelist returns the DeployerWhitelist param
func (k Keeper) DeployerWhitelist(ctx sdk.Context) (res []types.DeployerParams) {
	k.paramstore.Get(ctx, types.KeyDeployerWhitelist, &res)
	return
}

func (k Keeper) RollappsEnabled(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyRollappsEnabled, &res)
	return
}

func (k Keeper) HubExpectedBlockTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyHubExpectedBlockTime, &res)
	return
}

func (k Keeper) LivenessSlashTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyLivenessSlashTime, &res)
	return
}

func (k Keeper) LivenessSlashInterval(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyLivenessSlashInterval, &res)
	return
}

func (k Keeper) LivenessJailTime(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyLivenessJailTime, &res)
	return
}

func (k Keeper) IsAddressInDeployerWhiteList(ctx sdk.Context, address string) bool {
	whitelist := k.DeployerWhitelist(ctx)
	for _, item := range whitelist {
		if item.Address == address {
			return true
		}
	}
	return false
}
