package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// DistributeOnEpochEnd distributes coins from an array of gauges.
// It is called at the end of each epoch to distribute coins to the gauges that are active at that time.
func (k Keeper) DistributeOnEpochEnd(ctx sdk.Context, gauges []types.Gauge) (sdk.Coins, error) {
	cache := types.NewDenomLocksCache()

	const EpochEnd = true
	totalDistributedCoins, err := k.Distribute(ctx, gauges, cache, EpochEnd)
	if err != nil {
		return nil, fmt.Errorf("distribute gauges: %w", err)
	}

	// call post distribution hooks
	k.hooks.AfterEpochDistribution(ctx)

	k.checkFinishedGauges(ctx, gauges)

	return totalDistributedCoins, nil
}

// Distribute distributes coins from an array of gauges. It may be called either at the end or at the middle of
// the epoch. If it's called at the end, then the FilledEpochs field for every gauge is increased. Also, it uses
// a cache specific for asset gauges that helps reduce the number of x/lockup requests.
func (k Keeper) Distribute(ctx sdk.Context, gauges []types.Gauge, cache types.DenomLocksCache, epochEnd bool) (sdk.Coins, error) {
	// lockHolders is a map of address -> coins
	// it used as an aggregator for owners of the locks over all gauges
	lockHolders := NewRewardDistributionTracker()
	totalDistributedCoins := sdk.Coins{}

	// Get minimum distribution value from params
	minDistrValueCache := &DistributionValueCache{
		minDistrValue:      k.GetParams(ctx).MinValueForDistribution,
		denomToMinValueMap: make(map[string]math.Int),
	}

	for _, gauge := range gauges {
		var (
			gaugeDistributedCoins sdk.Coins
			err                   error
		)
		switch gauge.DistributeTo.(type) {
		case *types.Gauge_Asset:
			filteredLocks := k.GetDistributeToBaseLocks(ctx, gauge, cache) // get all locks that satisfy the gauge
			gaugeDistributedCoins, err = k.calculateAssetGaugeRewards(ctx, gauge, filteredLocks, &lockHolders, minDistrValueCache)
		case *types.Gauge_Rollapp:
			gaugeDistributedCoins, err = k.calculateRollappGaugeRewards(ctx, gauge, &lockHolders)
		case *types.Gauge_Endorsement:
			// endorsement gauges are distributed on epoch end
			// in that case, gaugeDistributedCoins == 0 as the rewards are distributed lazily during the epoch
			if epochEnd {
				err = k.updateEndorsementGaugeOnEpochEnd(ctx, gauge)
			}
		default:
			return nil, errorsmod.WithType(sdkerrors.ErrInvalidType, fmt.Errorf("gauge %d has an unsupported distribution type", gauge.Id))
		}
		if err != nil {
			return nil, err
		}

		if !gaugeDistributedCoins.Empty() {
			err = k.updateGaugePostDistribute(ctx, gauge, gaugeDistributedCoins, epochEnd)
			if err != nil {
				return nil, err
			}
			totalDistributedCoins = totalDistributedCoins.Add(gaugeDistributedCoins...)
		}
	}

	// apply the distribution to asset gauges
	err := k.distributeTrackedRewards(ctx, &lockHolders)
	if err != nil {
		return nil, err
	}

	return totalDistributedCoins, nil
}

// GetModuleToDistributeCoins returns sum of coins yet to be distributed for all of the module.
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getToDistributeCoinsFromGauges(k.getGaugesFromIterator(ctx, k.ActiveGaugesIterator(ctx)))
	upcomingGaugesDistr := k.getToDistributeCoinsFromGauges(k.getGaugesFromIterator(ctx, k.UpcomingGaugesIterator(ctx)))
	return activeGaugesDistr.Add(upcomingGaugesDistr...)
}

// GetModuleDistributedCoins returns sum of coins that have been distributed so far for all of the module.
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeGaugesDistr := k.getDistributedCoinsFromGauges(k.getGaugesFromIterator(ctx, k.ActiveGaugesIterator(ctx)))
	finishedGaugesDistr := k.getDistributedCoinsFromGauges(k.getGaugesFromIterator(ctx, k.FinishedGaugesIterator(ctx)))

	return activeGaugesDistr.Add(finishedGaugesDistr...)
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

// updateGaugePostDistribute increments the gauge's filled epochs field.
// Also adds the coins that were just distributed to the gauge's distributed coins field.
func (k Keeper) updateGaugePostDistribute(ctx sdk.Context, gauge types.Gauge, newlyDistributedCoins sdk.Coins, epochEnd bool) error {
	if epochEnd {
		gauge.FilledEpochs += 1
	}
	gauge.DistributedCoins = gauge.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setGauge(ctx, &gauge); err != nil {
		return err
	}
	return nil
}

// checkFinishedGauges checks if all non perpetual gauges provided have completed their required distributions.
// If complete, move the gauge from an active to a finished status.
func (k Keeper) checkFinishedGauges(ctx sdk.Context, gauges []types.Gauge) {
	for _, gauge := range gauges {
		if gauge.IsPerpetual {
			continue
		}

		// filled epoch is increased in this step and we compare with +1
		if gauge.NumEpochsPaidOver <= gauge.FilledEpochs+1 {
			if err := k.moveActiveGaugeToFinishedGauge(ctx, gauge); err != nil {
				panic(err)
			}
		}
	}
}
