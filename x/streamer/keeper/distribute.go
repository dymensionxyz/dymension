package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/utils/cache"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Distribute distributes rewards to the provided streams within provided epochs considering the max number
// of iterations.
// It also sends coins to the x/incentives module before the gauge distribution and emits an end block event.
// The method uses three caches:
//   - Stream cache for updating stream distributed coins
//   - Gauge cache for updating gauge coins
//   - Number of locks per denom to reduce the number of requests for x/lockup
//
// Returns distributed coins, the num of total iterations, and the error.
func (k Keeper) Distribute(
	ctx sdk.Context,
	epochPointers []types.EpochPointer,
	streams []types.Stream,
	maxIterations uint64,
	epochEnd bool,
) (coins sdk.Coins, iterations uint64, err error) {
	// Sort epoch pointers to distribute to shorter epochs first
	types.SortEpochPointers(epochPointers)

	// Total iterations counter
	totalIterations := uint64(0)

	// Init helper caches
	streamCache := cache.NewDistribution(streams, types.Stream.AddCoins, types.Stream.Key)
	gaugeCache := cache.NewDistribution(nil, incentivestypes.Gauge.AddCoins, incentivestypes.Gauge.Key)

	// Cache specific for asset gauges. Helps reduce the number of x/lockup requests.
	denomLockCache := incentivestypes.NewDenomLocksCache()

	for _, p := range epochPointers {
		if totalIterations >= maxIterations {
			break // no more iterations available for this block
		}

		remainIterations := maxIterations - totalIterations // always positive

		// Calculate rewards and fill caches
		newPointer, iters := k.CalculateRewards(ctx, p, remainIterations, streamCache, gaugeCache, denomLockCache)

		totalIterations += iters

		err = k.SaveEpochPointer(ctx, newPointer)
		if err != nil {
			return nil, 0, fmt.Errorf("save epoch pointer: %w", err)
		}
	}

	// Send coins to distribute to the x/incentives module
	err = k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, incentivestypes.ModuleName, streamCache.TotalDistrCoins())
	if err != nil {
		return nil, 0, fmt.Errorf("send coins: %w", err)
	}

	// Distribute the rewards
	_, err = k.ik.Distribute(ctx, gaugeCache.GetValues(), denomLockCache, epochEnd)
	if err != nil {
		return nil, 0, fmt.Errorf("distribute: %w", err)
	}

	// Save stream updates
	var rangeErr error
	streamCache.Range(func(stream types.Stream) bool {
		// If it is an epoch end, then update the stream info like the num of filled epochs
		if epochEnd {
			stream, rangeErr = k.UpdateStreamAtEpochEnd(ctx, stream)
			if rangeErr != nil {
				rangeErr = fmt.Errorf("update stream '%d' at epoch start: %w", stream.Id, rangeErr)
				return false
			}
		}
		rangeErr = k.SetStream(ctx, &stream)
		if rangeErr != nil {
			rangeErr = fmt.Errorf("set stream: %w", rangeErr)
			return false
		}
		return true
	})
	if rangeErr != nil {
		return nil, 0, rangeErr
	}

	return streamCache.TotalDistrCoins(), totalIterations, nil
}

// CalculateGaugeRewards calculates the rewards to be distributed for a specific gauge based on the provided
// coins, distribution record, and total weight. The method iterates through the coins, calculates the
// allocating amount based on the weight of the gauge and the total weight, and adds the allocated amount
// to the rewards. If the allocating amount is not positive, the coin is skipped.
func (k Keeper) CalculateGaugeRewards(ctx sdk.Context, coins sdk.Coins, record types.DistrRecord, totalWeight math.Int) (sdk.Coins, error) {
	if coins.Empty() {
		return nil, fmt.Errorf("coins to allocate cannot be empty")
	}

	if totalWeight.IsZero() {
		return nil, fmt.Errorf("distribution total weight cannot be zero")
	}

	weightDec := sdk.NewDecFromInt(record.Weight)
	totalDec := sdk.NewDecFromInt(totalWeight)
	rewards := sdk.NewCoins()

	for _, coin := range coins {
		if coin.IsZero() {
			continue
		}

		assetAmountDec := sdk.NewDecFromInt(coin.Amount)
		allocatingAmount := assetAmountDec.Mul(weightDec.Quo(totalDec)).TruncateInt()

		// when weight is too small and no amount is allocated, just skip this to avoid zero coin send issues
		if !allocatingAmount.IsPositive() {
			k.Logger(ctx).Info(fmt.Sprintf("allocating amount for gauge id '%d' with weight '%s' is not positive", record.GaugeId, record.Weight.String()))
			continue
		}

		allocatedCoin := sdk.Coin{Denom: coin.Denom, Amount: allocatingAmount}
		rewards = rewards.Add(allocatedCoin)
	}

	return rewards, nil
}

// CalculateRewards calculates rewards for streams and corresponding gauges. Is starts processing gauges from
// the specified pointer and considering the limit. This method doesn't have any state updates, it only
// calculates rewards and fills respective caches. Returns a new pointer and the num of iterations done.
func (k Keeper) CalculateRewards(
	ctx sdk.Context,
	pointer types.EpochPointer,
	limit uint64,
	streamCache *cache.Distribution[uint64, types.Stream],
	gaugeCache *cache.Distribution[uint64, incentivestypes.Gauge],
	denomLocksCache incentivestypes.DenomLocksCache,
) (newPointer types.EpochPointer, iterations uint64) {
	return IterateEpochPointer(pointer, streamCache.GetValues(), limit, func(v StreamGauge) (stop bool, weight uint64) {
		stream, found := streamCache.Get(v.Stream.Id)
		if !found {
			// this should never happen in practice since the initial cache contains all gauges we are iterating
			panic(fmt.Errorf("internal contract error: stream '%d' not found in the cache!", stream.Id))
		}

		// first, check the gauge in the cache
		gauge, found := gaugeCache.Get(v.Gauge.GaugeId)
		if !found {
			// validate the gauge exists
			var err error
			gauge, err = k.getActiveGaugeByID(ctx, v.Gauge.GaugeId)
			if err != nil {
				// we don't want to fail in this case, ignore this gauge
				k.Logger(ctx).
					With("gaugeID", v.Gauge.GaugeId, "error", err.Error()).
					Error("Can't distribute to gauge: failed to get active gauge")
				return false, 0 // continue, weight = 0, consider this operation as it is free
			}
			// add a new gauge to the cache
			gaugeCache.Add(gauge)
		}

		rewards, err := k.CalculateGaugeRewards(
			ctx,
			v.Stream.EpochCoins,
			v.Gauge,
			stream.DistributeTo.TotalWeight,
		)
		if err != nil {
			// we don't want to fail in this case, ignore this gauge
			k.Logger(ctx).
				With("streamID", stream.Id, "gaugeID", v.Gauge.GaugeId, "error", err.Error()).
				Error("Failed to distribute to gauge")
			return false, 0 // continue, weight = 0, consider this operation as it is free
		}

		// Update distributed coins for the stream
		streamCache.AddValueWithCoins(stream, rewards)
		gauge = gaugeCache.AddValueWithCoins(gauge, rewards)

		weight = k.getGaugeWeight(ctx, gauge, denomLocksCache)

		return false, weight
	})
}

// getActiveGaugeByID returns the active gauge with the given ID from the keeper.
// An error is returned if the gauge does not exist or if it is finished.
func (k Keeper) getActiveGaugeByID(ctx sdk.Context, gaugeID uint64) (incentivestypes.Gauge, error) {
	// validate the gauge exists
	gauge, err := k.ik.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return incentivestypes.Gauge{}, fmt.Errorf("get gauge by id '%d': %w", gaugeID, err)
	}
	// validate the gauge is not finished
	finished := gauge.IsFinishedGauge(ctx.BlockTime())
	if finished {
		return incentivestypes.Gauge{}, incentivestypes.UnexpectedFinishedGaugeError{GaugeId: gaugeID}
	}
	return *gauge, nil
}

// getGaugeWeight calculates the weight of a gauge based on its type.
// If the gauge is an asset gauge, the weight is equal to the number of associated lockups. Use cache to get this figure.
// If the gauge is a rollapp gauge, the weight is always 1. If the gauge has an unknown type, the weight is 0 as we
// assume that the gauge is always validated at this step and has a known type. The method also fills the cache under
// the hood.
func (k Keeper) getGaugeWeight(ctx sdk.Context, gauge incentivestypes.Gauge, cache incentivestypes.DenomLocksCache) uint64 {
	switch gauge.DistributeTo.(type) {
	case *incentivestypes.Gauge_Asset:
		// GetDistributeToBaseLocks fills the cache
		locks := k.ik.GetDistributeToBaseLocks(ctx, gauge, cache)
		// asset gauge weight is the num of associated lockups
		return uint64(len(locks))
	case *incentivestypes.Gauge_Rollapp:
		// rollapp gauge weight is always 1
		return 1
	default:
		// assume that the gauge is always validated at this step and has a known type
		return 0
	}
}
