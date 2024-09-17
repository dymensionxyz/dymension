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
	maxOperations uint64,
	epochEnd bool,
) (coins sdk.Coins, iterations uint64, err error) {
	// Sort epoch pointers to distribute to shorter epochs first. Our goal is to fill streams with
	// shorter epochs first. Otherwise, if a long stream has too many operations and takes entire blocks,
	// then we will never start processing shorter streams during the epoch, and only process them at the epoch end.
	types.SortEpochPointers(epochPointers)

	// Total operations counter. Each stream has some specific number of operations to process it. This counter
	// serves as a meter for the number of operations that have been performed during stream processing.
	// This is used to ensure that the method won't meet the upper bound for the complexity (maxOperations)
	// that this method is capable of.
	totalOperations := uint64(0)
	totalDistributed := sdk.NewCoins()

	// Init helper caches
	streamCache := cache.NewInsertionOrdered(types.Stream.Key, streams...)
	gaugeCache := cache.NewInsertionOrdered(incentivestypes.Gauge.Key)

	// Cache specific for asset gauges. Helps reduce the number of x/lockup requests.
	denomLockCache := incentivestypes.NewDenomLocksCache()

	for _, p := range epochPointers {
		if totalOperations >= maxOperations {
			// The upped bound of operations is met. No more operations available for this block.
			break
		}

		remainOperations := maxOperations - totalOperations // always positive

		// Calculate rewards and fill caches
		distrCoins, newPointer, iters := k.CalculateRewards(ctx, p, remainOperations, streamCache, gaugeCache, denomLockCache)

		totalOperations += iters
		totalDistributed = totalDistributed.Add(distrCoins...)

		err = k.SaveEpochPointer(ctx, newPointer)
		if err != nil {
			return nil, 0, fmt.Errorf("save epoch pointer: %w", err)
		}
	}

	// Send coins to distribute to the x/incentives module
	if !totalDistributed.Empty() {
		err = k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, incentivestypes.ModuleName, totalDistributed)
		if err != nil {
			return nil, 0, fmt.Errorf("send coins: %w", err)
		}
	}

	// Distribute the rewards
	_, err = k.ik.Distribute(ctx, gaugeCache.GetAll(), denomLockCache, epochEnd)
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
				rangeErr = fmt.Errorf("update stream at epoch start: stream %d: %w", stream.Id, rangeErr)
				return true
			}
		}
		rangeErr = k.SetStream(ctx, &stream)
		if rangeErr != nil {
			rangeErr = fmt.Errorf("set stream: %w", rangeErr)
			return true
		}
		return false
	})
	if rangeErr != nil {
		return nil, 0, rangeErr
	}

	return totalDistributed, totalOperations, nil
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
			k.Logger(ctx).Info(fmt.Sprintf("allocating amount for gauge is not positive: gauge '%d', weight %s", record.GaugeId, record.Weight.String()))
			continue
		}

		allocatedCoin := sdk.Coin{Denom: coin.Denom, Amount: allocatingAmount}
		rewards = rewards.Add(allocatedCoin)
	}

	return rewards, nil
}

// CalculateRewards calculates rewards for streams and corresponding gauges. Is starts processing gauges from
// the specified pointer and considering the limit. This method doesn't have any state updates, it only
// calculates rewards and fills respective caches.
// Returns a new pointer, total distr coins, and the num of iterations.
func (k Keeper) CalculateRewards(
	ctx sdk.Context,
	pointer types.EpochPointer,
	limit uint64,
	streamCache *cache.InsertionOrdered[uint64, types.Stream],
	gaugeCache *cache.InsertionOrdered[uint64, incentivestypes.Gauge],
	denomLocksCache incentivestypes.DenomLocksCache,
) (distributedCoins sdk.Coins, newPointer types.EpochPointer, operations uint64) {
	distributedCoins = sdk.NewCoins()
	pointer, operations = IterateEpochPointer(pointer, streamCache.GetAll(), limit, func(v StreamGauge) (stop bool, operations uint64) {
		// get stream from the cache since we need to use the last updated version
		stream := streamCache.MustGet(v.Stream.Id)

		// get gauge from the cache since we need to use the last updated version
		gauge, ok := gaugeCache.Get(v.Gauge.GaugeId)
		if !ok {
			// If the gauge is not found, then
			// 1. Request it from the incentives keeper and validate the gauge exists
			// 2. Validate that it's not finished
			// 3. If everything is fine, then add the gauge to the cache, and use the cached versions in the future
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
			gaugeCache.Upsert(gauge)
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

		// update distributed coins for the stream
		stream.AddDistributedCoins(rewards)
		streamCache.Upsert(stream)

		// update distributed coins for the gauge
		gauge.AddCoins(rewards)
		gaugeCache.Upsert(gauge)

		// get gauge weight and update denomLocksCache under the hood
		operations = k.getGaugeLockNum(ctx, gauge, denomLocksCache)

		distributedCoins = distributedCoins.Add(rewards...)

		return false, operations
	})
	return distributedCoins, pointer, operations
}

// getActiveGaugeByID returns the active gauge with the given ID from the keeper.
// An error is returned if the gauge does not exist or if it is finished.
func (k Keeper) getActiveGaugeByID(ctx sdk.Context, gaugeID uint64) (incentivestypes.Gauge, error) {
	// validate the gauge exists
	gauge, err := k.ik.GetGaugeByID(ctx, gaugeID)
	if err != nil {
		return incentivestypes.Gauge{}, fmt.Errorf("get gauge: id %d: %w", gaugeID, err)
	}
	// validate the gauge is not finished
	finished := gauge.IsFinishedGauge(ctx.BlockTime())
	if finished {
		return incentivestypes.Gauge{}, incentivestypes.UnexpectedFinishedGaugeError{GaugeId: gaugeID}
	}
	return *gauge, nil
}

// getGaugeLockNum returns the number of locks for the specified gauge.
// If the gauge is an asset gauge, return the number of associated lockups. Use cache to get this figure.
// If the gauge is a rollapp gauge, the number is always 1. Imagine a rollapp as a single lockup.
// If the gauge has an unknown type, the weight is 0 as we assume that the gauge is always validated at this step
// and has a known type. The method also fills the cache under the hood.
func (k Keeper) getGaugeLockNum(ctx sdk.Context, gauge incentivestypes.Gauge, cache incentivestypes.DenomLocksCache) uint64 {
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
