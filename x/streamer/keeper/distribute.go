package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

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

type DistributeRewardsResult struct {
	NewPointer types.EpochPointer
	Iterations uint64
}

// CalculateRewards calculates rewards for streams and corresponding gauges. Is starts processing gauges from
// the specified pointer and considering the limit. This method doesn't have any state updates and validations
// (for example, if the gauge exists or is unfinished), it only calculates rewards and fills respective caches.
func (k Keeper) CalculateRewards(
	ctx sdk.Context,
	pointer types.EpochPointer,
	limit uint64,
	streamInfo *streamInfo,
	gaugeInfo *gaugeInfo,
	denomLocksCache incentivestypes.DenomLocksCache,
) DistributeRewardsResult {
	newPointer, iterations := IterateEpochPointer(pointer, streamInfo.getStreams(), limit, func(v StreamGauge) (stop bool, weight uint64) {
		// validate the gauge exists
		gauge, err := k.getActiveGaugeByID(ctx, v.Gauge.GaugeId)
		if err != nil {
			// we don't want to fail in this case, ignore this gauge
			k.Logger(ctx).
				With("gaugeID", v.Gauge.GaugeId, "error", err.Error()).
				Error("Can't distribute to gauge: failed to get active gauge")
			return false, 0 // continue, weight = 0, consider this operation as it is free
		}

		rewards, err := k.CalculateGaugeRewards(
			ctx,
			v.Stream.EpochCoins,
			v.Gauge,
			v.Stream.DistributeTo.TotalWeight,
		)
		if err != nil {
			// we don't want to fail in this case, ignore this gauge
			k.Logger(ctx).
				With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", err.Error()).
				Error("Failed to distribute to gauge")
			return false, 0 // continue, weight = 0, consider this operation as it is free
		}

		// Update distributed coins for the stream
		streamInfo.addDistrCoins(v.Stream, rewards)
		gaugeInfo.addDistrCoins(gauge, rewards)

		weight = k.getGaugeWeight(ctx, gauge, denomLocksCache)

		return false, weight
	})

	return DistributeRewardsResult{
		NewPointer: newPointer,
		Iterations: iterations,
	}
}

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
