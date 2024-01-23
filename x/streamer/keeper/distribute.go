package keeper

import (
	"fmt"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DistributeByWeights allocates and distributes coin according a gauge’s proportional weight that is recorded in the record.
func (k Keeper) DistributeByWeights(ctx sdk.Context, coins sdk.Coins, distrInfo *types.DistrInfo) (sdk.Coins, error) {
	logger := k.Logger(ctx)

	if coins.Empty() {
		return coins, fmt.Errorf("coins to allocate cannot be empty")
	}

	if distrInfo.TotalWeight.IsZero() {
		return sdk.Coins{}, fmt.Errorf("distribution total weight cannot be zero")
	}

	totalDistrCoins := sdk.NewCoins()
	totalWeightDec := sdk.NewDecFromInt(distrInfo.TotalWeight)
	for _, coin := range coins {
		if coin.IsZero() {
			continue
		}
		assetAmountDec := sdk.NewDecFromInt(coin.Amount)
		for _, record := range distrInfo.Records {
			allocatingAmount := assetAmountDec.Mul(sdk.NewDecFromInt(record.Weight).Quo(totalWeightDec)).TruncateInt()

			// when weight is too small and no amount is allocated, just skip this to avoid zero coin send issues
			if !allocatingAmount.IsPositive() {
				logger.Info(fmt.Sprintf("allocating amount for (%d, %s) record is not positive", record.GaugeId, record.Weight.String()))
				continue
			}

			_, err := k.ik.GetGaugeByID(ctx, record.GaugeId)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to get gauge %d", record.GaugeId), "error", err.Error())
				continue
			}

			allocatedCoin := sdk.Coin{Denom: coin.Denom, Amount: allocatingAmount}
			err = k.ik.AddToGaugeRewards(ctx, k.ak.GetModuleAddress(types.ModuleName), sdk.NewCoins(allocatedCoin), record.GaugeId)
			if err != nil {
				logger.Error("failed to add to gauge rewards", "error", err.Error())
				continue
			}
			totalDistrCoins = totalDistrCoins.Add(allocatedCoin)
		}
	}

	return totalDistrCoins, nil
}

// Distribute distributes coins from an array of streams to all eligible locks.
func (k Keeper) Distribute(ctx sdk.Context, streams []types.Stream) (sdk.Coins, error) {
	totalDistributedCoins := sdk.Coins{}
	for _, stream := range streams {
		streamDistributedCoins, err := k.distributeStream(ctx, stream)
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
	totalDistrCoins := sdk.NewCoins()
	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := uint64(stream.NumEpochsPaidOver - stream.FilledEpochs)

	for _, coin := range remainCoins {
		epochAmt := coin.Amount.Quo(sdk.NewInt(int64(remainEpochs)))
		if epochAmt.IsPositive() {
			totalDistrCoins = totalDistrCoins.Add(sdk.Coin{Denom: coin.Denom, Amount: epochAmt})
		}
	}

	totalDistrCoins, err := k.DistributeByWeights(ctx, totalDistrCoins, stream.DistributeTo)
	if err != nil {
		return nil, err
	}

	err = k.updateStreamPostDistribute(ctx, stream, totalDistrCoins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtDistribution,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
			sdk.NewAttribute(types.AttributeAmount, totalDistrCoins.String()),
		),
	})
	return totalDistrCoins, nil
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
