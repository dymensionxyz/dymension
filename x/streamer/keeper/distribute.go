package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// DistributeByWeights allocates and distributes coin according a gauge’s proportional weight that is recorded in the record.
func (k Keeper) DistributeByWeights(ctx sdk.Context, coins sdk.Coins, distrInfo *types.DistrInfo) (sdk.Coins, error) {
	if coins.Empty() {
		return coins, fmt.Errorf("coins to allocate cannot be empty")
	}

	if distrInfo.TotalWeight.IsZero() {
		return sdk.Coins{}, fmt.Errorf("distribution total weight cannot be zero")
	}

	totalDistrCoins := sdk.NewCoins()
	for _, record := range distrInfo.Records {
		allocatedCoins, err := k.DistributeToGauge(ctx, coins, record, distrInfo.TotalWeight)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("distribute to gauge %d: %w", record.GaugeId, err)
		}
		totalDistrCoins = totalDistrCoins.Add(allocatedCoins...)
	}

	return totalDistrCoins, nil
}

func (k Keeper) DistributeToGauge(ctx sdk.Context, coins sdk.Coins, record types.DistrRecord, totalWeight math.Int) (sdk.Coins, error) {
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

// Distribute distributes coins from an array of streams to all eligible locks.
func (k Keeper) Distribute(ctx sdk.Context, streams []types.Stream) (sdk.Coins, error) {
	totalDistributedCoins := sdk.Coins{}
	streamDistributedCoins := sdk.Coins{}
	for _, stream := range streams {
		wrappedDistributeFn := func(ctx sdk.Context) error {
			var err error
			streamDistributedCoins, err = k.distributeStream(ctx, stream)
			return err
		}

		err := osmoutils.ApplyFuncIfNoError(ctx, wrappedDistributeFn)
		if err != nil {
			ctx.Logger().Error("Failed to distribute stream", "streamID", stream.Id, "error", err.Error())
			continue
		}
		totalDistributedCoins = totalDistributedCoins.Add(streamDistributedCoins...)
	}

	return totalDistributedCoins, nil
}

// distributeStream runs the distribution logic for a stream, and adds the sends to
// the distrInfo struct. It also updates the stream for the distribution.
func (k Keeper) distributeStream(ctx sdk.Context, stream types.Stream) (sdk.Coins, error) {
	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := stream.NumEpochsPaidOver - stream.FilledEpochs
	epochCoins := remainCoins.QuoInt(math.NewIntFromUint64(remainEpochs))

	// If the stream uses a sponsorship plan, query it and update stream distr info. The distribution
	// might be empty and this is a valid scenario. In that case, we'll just skip at without
	// filling the epoch.
	if stream.Sponsored {
		distr, err := k.sk.GetDistribution(ctx)
		if err != nil {
			return sdk.Coins{}, fmt.Errorf("failed to get sponsorship distribution: %w", err)
		}
		info := types.DistrInfoFromDistribution(distr)
		// Update stream distr info
		stream.DistributeTo = info
	}

	distributedCoins, err := k.DistributeByWeights(ctx, epochCoins, stream.DistributeTo)
	if err != nil {
		return nil, err
	}

	err = k.updateStreamPostDistribute(ctx, stream, distributedCoins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtDistribution,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
			sdk.NewAttribute(types.AttributeAmount, distributedCoins.String()),
		),
	})

	return distributedCoins, nil
}

// updateStreamPostDistribute increments the stream's filled epochs field.
// Also adds the coins that were just distributed to the stream's distributed coins field.
func (k Keeper) updateStreamPostDistribute(ctx sdk.Context, stream types.Stream, newlyDistributedCoins sdk.Coins) error {
	stream.FilledEpochs += 1
	stream.DistributedCoins = stream.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setStream(ctx, &stream); err != nil {
		return err
	}

	// Check if stream has completed its distribution
	if stream.FilledEpochs >= stream.NumEpochsPaidOver {
		if err := k.moveActiveStreamToFinishedStream(ctx, stream); err != nil {
			return err
		}
	}

	return nil
}
