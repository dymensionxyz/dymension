package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/internal/pagination"
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
		err = k.ik.AddToGaugeRewards(ctx, k.ak.GetModuleAddress(types.ModuleName), sdk.NewCoins(allocatedCoin), record.GaugeId)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("add rewards to gauge %d: %w", record.GaugeId, err)
		}

		totalAllocated = totalAllocated.Add(allocatedCoin)
	}

	return totalAllocated, nil
}

// UpdateStreamAtEpochStart updates the stream for a new epoch. Streams distribute rewards post factum.
// Meaning, first increase the filled epoch pointer, then distribute rewards for this epoch.
func (k Keeper) UpdateStreamAtEpochStart(ctx sdk.Context, stream types.Stream) (types.Stream, error) {
	// Check if stream has completed its distribution. This is a post factum check.
	if stream.FilledEpochs >= stream.NumEpochsPaidOver {
		err := k.moveActiveStreamToFinishedStream(ctx, stream)
		if err != nil {
			return types.Stream{}, fmt.Errorf("move active stream to finished stream: %w", err)
		}
		return stream, nil
	}

	// If the stream is not finalized, update it for the next distribution

	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := stream.NumEpochsPaidOver - stream.FilledEpochs
	epochCoins := remainCoins.QuoInt(math.NewIntFromUint64(remainEpochs))

	// If the stream uses a sponsorship plan, query it and update stream distr info. The distribution
	// might be empty and this is a valid scenario. In that case, we'll just skip at without
	// filling the epoch.
	if stream.Sponsored {
		distr, err := k.sk.GetDistribution(ctx)
		if err != nil {
			return types.Stream{}, fmt.Errorf("get sponsorship distribution: %w", err)
		}
		// Update stream distr info
		stream.DistributeTo = types.DistrInfoFromDistribution(distr)
	}

	// Add coins to distribute during the next epoch
	stream.EpochCoins = epochCoins

	// Don't fill streams in which there's nothing to fill. Note that rewards are distributed post factum.
	// I.e., first increase the filled epoch number, then distribute rewards during the epoch.
	if !stream.DistributeTo.TotalWeight.IsZero() {
		stream.FilledEpochs += 1
	}

	// TODO: can we delete this event?
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtDistribution,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
			sdk.NewAttribute(types.AttributeAmount, epochCoins.String()),
		),
	})

	return stream, nil
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
		distributed, errX := k.DistributeToGauge(ctx, v.Stream.EpochCoins, v.Gauge, v.Stream.DistributeTo.TotalWeight)
		if errX != nil {
			// Ignore this gauge
			k.Logger(ctx).
				With("streamID", v.Stream.Id, "gaugeID", v.Gauge.GaugeId, "error", errX.Error()).
				Error("Failed to distribute to gauge")
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
