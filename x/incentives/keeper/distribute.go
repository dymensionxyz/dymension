package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Distribute distributes coins from an array of gauges.
// It is called at the end of each epoch to distribute coins to the gauges that are active at that time.
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	lockHolders := newDistributionInfo()

	totalDistributedCoins := sdk.Coins{}
	for _, gauge := range gauges {
		var (
			gaugeDistributedCoins sdk.Coins
			err                   error
		)
		switch gauge.DistributeTo.(type) {
		case *types.Gauge_Asset:
			gaugeDistributedCoins, err = k.distributeToAssetGauge(ctx, gauge, &lockHolders)
		case *types.Gauge_Rollapp:
			gaugeDistributedCoins, err = k.distributeToRollappGauge(ctx, gauge)
		default:
			return nil, fmt.Errorf("gauge %d has an unsupported distribution type", gauge.Id)
		}
		if err != nil {
			return nil, err
		}

		totalDistributedCoins = totalDistributedCoins.Add(gaugeDistributedCoins...)
	}

	// apply the distribution to asset gauges
	err := k.doDistributionSends(ctx, &lockHolders)
	if err != nil {
		return nil, err
	}
	k.hooks.AfterEpochDistribution(ctx)

	k.checkFinishDistribution(ctx, gauges)
	return totalDistributedCoins, nil
}

// getDistributedCoinsFromGauges returns coins that have been distributed already from the provided gauges
func (k Keeper) getDistributedCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	for _, gauge := range gauges {
		coins = coins.Add(gauge.DistributedCoins...)
	}
	return coins
}

// getToDistributeCoinsFromGauges returns coins that have not been distributed yet from the provided gauges
func (k Keeper) getToDistributeCoinsFromGauges(gauges []types.Gauge) sdk.Coins {
	coins := sdk.Coins{}
	distributed := sdk.Coins{}

	for _, gauge := range gauges {
		coins = coins.Add(gauge.Coins...)
		distributed = distributed.Add(gauge.DistributedCoins...)
	}
	return coins.Sub(distributed...)
}

// moveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) moveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge types.Gauge) error {
	// validation for current time and distribution start time
	if ctx.BlockTime().Before(gauge.StartTime) {
		return fmt.Errorf("gauge is not able to start distribution yet: %s >= %s", ctx.BlockTime().String(), gauge.StartTime.String())
	}

	timeKey := getTimeKey(gauge.StartTime)
	if err := k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	return nil
}

// moveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge types.Gauge) error {
	timeKey := getTimeKey(gauge.StartTime)
	if err := k.deleteGaugeRefByKey(ctx, combineKeys(types.KeyPrefixActiveGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	if err := k.addGaugeRefByKey(ctx, combineKeys(types.KeyPrefixFinishedGauges, timeKey), gauge.Id); err != nil {
		return err
	}
	assetGauge, ok := gauge.DistributeTo.(*types.Gauge_Asset)
	if ok {
		if err := k.deleteGaugeIDForDenom(ctx, gauge.Id, assetGauge.Asset.Denom); err != nil {
			return err
		}
	}
	k.hooks.AfterFinishDistribution(ctx, gauge.Id)
	return nil
}

// updateGaugePostDistribute increments the gauge's filled epochs field.
// Also adds the coins that were just distributed to the gauge's distributed coins field.
func (k Keeper) updateGaugePostDistribute(ctx sdk.Context, gauge types.Gauge, newlyDistributedCoins sdk.Coins) error {
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}
	return nil
}

// checkFinishDistribution checks if all non perpetual gauges provided have completed their required distributions.
// If complete, move the gauge from an active to a finished status.
func (k Keeper) checkFinishDistribution(ctx sdk.Context, gauges []types.Gauge) {
	for _, gauge := range gauges {
		// filled epoch is increased in this step and we compare with +1
		if !gauge.IsPerpetual && gauge.NumEpochsPaidOver <= gauge.FilledEpochs+1 {
			if err := k.moveActiveGaugeToFinishedGauge(ctx, gauge); err != nil {
				panic(err)
			}
		}
	}
}
