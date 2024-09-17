package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/utils/pagination"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (k Keeper) DistributeToGauge(ctx sdk.Context, coins sdk.Coins, record types.DistrRecord, totalWeight math.Int) (sdk.Coins, error) {
	if coins.Empty() {
		return sdk.Coins{}, fmt.Errorf("coins to allocate cannot be empty")
	}

	if totalWeight.IsZero() {
		return sdk.Coins{}, fmt.Errorf("distribution total weight cannot be zero")
	}

	totalAllocated := sdk.NewCoins()
	for _, coin := range coins {
		if coin.IsZero() {
			continue
		}

		assetAmountDec := sdk.NewDecFromInt(coin.Amount)
		weightDec := sdk.NewDecFromInt(record.Weight)
		totalDec := sdk.NewDecFromInt(totalWeight)
		allocatingAmount := assetAmountDec.Mul(weightDec.Quo(totalDec)).TruncateInt()

		// when weight is too small and no amount is allocated, just skip this to avoid zero coin send issues
		if !allocatingAmount.IsPositive() {
			k.Logger(ctx).Info(fmt.Sprintf("allocating amount for (%d, %s) record is not positive", record.GaugeId, record.Weight.String()))
			continue
		}

		_, err := k.ik.GetGaugeByID(ctx, record.GaugeId)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("get gauge %d: %w", record.GaugeId, err)
		}

		allocatedCoin := sdk.Coin{Denom: coin.Denom, Amount: allocatingAmount}
		err = k.ik.AddToGaugeRewardsByID(ctx, k.ak.GetModuleAddress(types.ModuleName), sdk.NewCoins(allocatedCoin), record.GaugeId)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("add rewards to gauge %d: %w", record.GaugeId, err)
		}

		totalAllocated = totalAllocated.Add(allocatedCoin)
	}

	return totalAllocated, nil
}

type DistributeRewardsResult struct {
	NewPointer       types.EpochPointer
	FilledStreams    []types.Stream
	DistributedCoins sdk.Coins
	Iterations       uint64
}

// DistributeRewards distributes all streams rewards to the corresponding gauges starting with
// the specified pointer and considering the limit.
func (k Keeper) DistributeRewards(
	ctx sdk.Context,
	pointer types.EpochPointer,
	limit uint64,
	streams []types.Stream,
) DistributeRewardsResult {
	totalDistributed := sdk.NewCoins()

	// Temporary map for convenient calculations
	streamUpdates := make(map[uint64]sdk.Coins, len(streams))

	// Distribute to all the remaining gauges that are left after EndBlock
	newPointer, iterations := IterateEpochPointer(pointer, streams, limit, func(v StreamGauge) pagination.Stop {
		var distributed sdk.Coins
		err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			var err error
			distributed, err = k.DistributeToGauge(ctx, v.Stream.EpochCoins, v.Gauge, v.Stream.DistributeTo.TotalWeight)
			return err
		})
		if err != nil {
			// Ignore this gauge
			k.Logger(ctx).
				With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", err.Error()).
				Error("Failed to distribute to gauge")
			return pagination.Continue
		}

		totalDistributed = totalDistributed.Add(distributed...)

		// Update distributed coins for the stream
		update := streamUpdates[v.Stream.Id]
		update = update.Add(distributed...)
		streamUpdates[v.Stream.Id] = update

		return pagination.Continue
	})

	for i, s := range streams {
		s.DistributedCoins = s.DistributedCoins.Add(streamUpdates[s.Id]...)
		streams[i] = s
	}

	return DistributeRewardsResult{
		NewPointer:       newPointer,
		FilledStreams:    streams, // Make sure that the returning slice is always sorted
		DistributedCoins: totalDistributed,
		Iterations:       iterations,
	}
}
