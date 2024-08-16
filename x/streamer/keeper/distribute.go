package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

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

// Distribute distributes coins from an array of streams to all eligible locks.
// TODO: delete
func (k Keeper) Distribute(ctx sdk.Context, streams []types.Stream) (sdk.Coins, error) {
	totalDistributedCoins := sdk.Coins{}
	streamDistributedCoins := sdk.Coins{}
	for _, stream := range streams {
		wrappedDistributeFn := func(ctx sdk.Context) error {
			var err error
			err = k.UpdateStreamAtEpochStart(ctx, stream)
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

// UpdateStreamAtEpochStart updates the stream for a new epoch.
func (k Keeper) UpdateStreamAtEpochStart(ctx sdk.Context, stream types.Stream) error {
	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := stream.NumEpochsPaidOver - stream.FilledEpochs
	epochCoins := remainCoins.QuoInt(math.NewIntFromUint64(remainEpochs))

	// If the stream uses a sponsorship plan, query it and update stream distr info. The distribution
	// might be empty and this is a valid scenario. In that case, we'll just skip at without
	// filling the epoch.
	if stream.Sponsored {
		distr, err := k.sk.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get sponsorship distribution: %w", err)
		}
		// Update stream distr info
		stream.DistributeTo = types.DistrInfoFromDistribution(distr)
	}

	// Don't fill streams in which there's nothing to fill
	if !stream.DistributeTo.TotalWeight.IsZero() {
		stream.FilledEpochs += 1
	}
	stream.EpochCoins = epochCoins

	err := k.setStream(ctx, &stream)
	if err != nil {
		return fmt.Errorf("set stream: %w", err)
	}

	// Check if stream has completed its distribution
	if stream.FilledEpochs >= stream.NumEpochsPaidOver {
		err := k.moveActiveStreamToFinishedStream(ctx, stream)
		if err != nil {
			return fmt.Errorf("move active stream to finished stream: %w", err)
		}
	}

	// TODO: can we delete this event?
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtDistribution,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
			sdk.NewAttribute(types.AttributeAmount, epochCoins.String()),
		),
	})

	return nil
}
