package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateRollappGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateRollappGauge(ctx sdk.Context, rollappId string) (uint64, error) {
	// Ensure the rollapp exists
	_, found := k.rk.GetRollapp(ctx, rollappId)
	if !found {
		return 0, fmt.Errorf("rollapp %s not found", rollappId)
	}

	gauge := types.NewRollappGauge(k.GetLastGaugeID(ctx)+1, rollappId)

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.SetLastGaugeID(ctx, gauge.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, true)
	if err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

// calculateRollappGaugeRewards computes the reward distribution for a rollapp gauge.
// Returns the total coins allocated for distribution.
func (k Keeper) calculateRollappGaugeRewards(ctx sdk.Context, gauge types.Gauge, tracker *RewardDistributionTracker) (sdk.Coins, error) {
	rollapp, found := k.rk.GetRollapp(ctx, gauge.GetRollapp().RollappId)
	if !found {
		return sdk.Coins{}, fmt.Errorf("gauge %d: rollapp %s not found", gauge.Id, gauge.GetRollapp().RollappId)
	}

	// if we are in active only mode, endorse only active rollapps
	// - proposer is not sentinel
	// - proposer is not kickable (the rollapp is active)
	if k.RollappGaugesMode(ctx) == types.Params_ActiveOnly {
		proposer := k.sk.GetProposer(ctx, gauge.GetRollapp().RollappId)
		if proposer.Sentinel() {
			ctx.Logger().Debug("rollapp is not active, skipping")
			return sdk.Coins{}, nil
		}
		if k.sk.Kickable(ctx, proposer) {
			ctx.Logger().Debug("proposer is kickable, rollapp is not endorsed")
			return sdk.Coins{}, nil
		}
	}

	totalDistrCoins := gauge.Coins.Sub(gauge.DistributedCoins...) // distribute all remaining coins
	if totalDistrCoins.Empty() {
		ctx.Logger().Debug(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		return sdk.Coins{}, nil
	}

	// Add rewards to the tracker for the rollapp owner
	owner := rollapp.Owner
	err := tracker.addLockRewards(owner, gauge.Id, totalDistrCoins)
	if err != nil {
		return sdk.Coins{}, err
	}

	return totalDistrCoins, nil
}
