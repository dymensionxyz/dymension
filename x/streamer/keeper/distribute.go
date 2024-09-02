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
) DistributeRewardsResult {
	newPointer, iterations := IterateEpochPointer(pointer, streamInfo.getStreams(), limit, func(v StreamGauge) (stop bool, weight uint64) {
		added, err := k.CalculateGaugeRewards(
			ctx,
			v.Stream.EpochCoins,
			v.Gauge,
			v.Stream.DistributeTo.TotalWeight,
		)
		if err != nil {
			// Ignore this gauge
			k.Logger(ctx).
				With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", err.Error()).
				Error("Failed to distribute to gauge")
			return false, 0 // weight = 0, consider this operation as it is free
		}

		// Update distributed coins for the stream
		streamInfo.addDistrCoins(v.Stream, added)
		gaugeInfo.addDistrCoins(v.Gauge.GaugeId, added)

		return false, 1
	})

	return DistributeRewardsResult{
		NewPointer: newPointer,
		Iterations: iterations,
	}
}
