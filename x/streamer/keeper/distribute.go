package keeper

import (
	"fmt"

	db "github.com/tendermint/tm-db"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getDistributedCoinsFromStreams returns coins that have been distributed already from the provided streams
func (k Keeper) getDistributedCoinsFromStreams(streams []types.Stream) sdk.Coins {
	coins := sdk.Coins{}
	for _, stream := range streams {
		coins = coins.Add(stream.DistributedCoins...)
	}
	return coins
}

// getToDistributeCoinsFromStreams returns coins that have not been distributed yet from the provided streams
func (k Keeper) getToDistributeCoinsFromStreams(streams []types.Stream) sdk.Coins {
	coins := sdk.Coins{}
	distributed := sdk.Coins{}

	for _, stream := range streams {
		coins = coins.Add(stream.Coins...)
		distributed = distributed.Add(stream.DistributedCoins...)
	}
	return coins.Sub(distributed...)
}

// getToDistributeCoinsFromIterator utilizes iterator to return a list of streams.
// From these streams, coins that have not yet been distributed are returned
func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromStreams(k.getStreamsFromIterator(ctx, iterator))
}

// getDistributedCoinsFromIterator utilizes iterator to return a list of streams.
// From these streams, coins that have already been distributed are returned
func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromStreams(k.getStreamsFromIterator(ctx, iterator))
}

// moveUpcomingStreamToActiveStream moves a stream that has reached it's start time from an upcoming to an active status.
func (k Keeper) moveUpcomingStreamToActiveStream(ctx sdk.Context, stream types.Stream) error {
	// validation for current time and distribution start time
	if ctx.BlockTime().Before(stream.StartTime) {
		return fmt.Errorf("stream is not able to start distribution yet: %s >= %s", ctx.BlockTime().String(), stream.StartTime.String())
	}

	timeKey := getTimeKey(stream.StartTime)
	if err := k.deleteStreamRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingStreams, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.addStreamRefByKey(ctx, combineKeys(types.KeyPrefixActiveStreams, timeKey), stream.Id); err != nil {
		return err
	}
	return nil
}

// moveActiveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveStreamToFinishedStream(ctx sdk.Context, stream types.Stream) error {
	timeKey := getTimeKey(stream.StartTime)
	if err := k.deleteStreamRefByKey(ctx, combineKeys(types.KeyPrefixActiveStreams, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.addStreamRefByKey(ctx, combineKeys(types.KeyPrefixFinishedStreams, timeKey), stream.Id); err != nil {
		return err
	}
	for _, coin := range stream.Coins {
		if err := k.deleteStreamIDForDenom(ctx, stream.Id, coin.Denom); err != nil {
			return err
		}
	}
	return nil
}

// distributeInternal runs the distribution logic for a stream, and adds the sends to
// the distrInfo struct. It also updates the stream for the distribution.
func (k Keeper) distributeInternal(
	ctx sdk.Context, stream types.Stream,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := uint64(stream.NumEpochsPaidOver - stream.FilledEpochs)

	for _, coin := range remainCoins {
		epochAmt := coin.Amount.Quo(sdk.NewInt(int64(remainEpochs)))
		if epochAmt.IsPositive() {
			newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: epochAmt}
			totalDistrCoins = totalDistrCoins.Add(newlyDistributedCoin)
		}
	}
	totalDistrCoins = totalDistrCoins.Sort()

	err := k.updateStreamPostDistribute(ctx, stream, totalDistrCoins)
	if err != nil {
		return nil, err
	}

	distAddr, err := sdk.AccAddressFromBech32(stream.DistributeTo)
	if err != nil {
		return nil, err
	}

	err = k.bk.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		distAddr,
		totalDistrCoins,
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtDistribution,
			sdk.NewAttribute(types.AttributeReceiver, distAddr.String()),
			sdk.NewAttribute(types.AttributeAmount, totalDistrCoins.String()),
		),
	})
	return nil, nil
}

// updateStreamPostDistribute increments the stream's filled epochs field.
// Also adds the coins that were just distributed to the stream's distributed coins field.
func (k Keeper) updateStreamPostDistribute(ctx sdk.Context, stream types.Stream, newlyDistributedCoins sdk.Coins) error {
	stream.FilledEpochs += 1
	stream.DistributedCoins = stream.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setStream(ctx, &stream); err != nil {
		return err
	}
	return nil
}

// Distribute distributes coins from an array of streams to all eligible locks.
func (k Keeper) Distribute(ctx sdk.Context, streams []types.Stream) (sdk.Coins, error) {
	totalDistributedCoins := sdk.Coins{}
	for _, stream := range streams {
		streamDistributedCoins, err := k.distributeInternal(ctx, stream)
		if err != nil {
			return nil, err
		}
		totalDistributedCoins = totalDistributedCoins.Add(streamDistributedCoins...)
	}

	k.checkFinishDistribution(ctx, streams)
	return totalDistributedCoins, nil
}

// checkFinishDistribution checks if all non perpetual streams provided have completed their required distributions.
// If complete, move the stream from an active to a finished status.
func (k Keeper) checkFinishDistribution(ctx sdk.Context, streams []types.Stream) {
	for _, stream := range streams {
		// filled epoch is increased in this step and we compare with +1
		if stream.NumEpochsPaidOver <= stream.FilledEpochs+1 {
			if err := k.moveActiveStreamToFinishedStream(ctx, stream); err != nil {
				panic(err)
			}
		}
	}
}

// GetModuleToDistributeCoins returns sum of coins yet to be distributed for all of the module.
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activeStreamsDistr := k.getToDistributeCoinsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
	upcomingStreamsDistr := k.getToDistributeCoinsFromIterator(ctx, k.UpcomingStreamsIterator(ctx))
	return activeStreamsDistr.Add(upcomingStreamsDistr...)
}

// GetModuleDistributedCoins returns sum of coins that have been distributed so far for all of the module.
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeStreamsDistr := k.getDistributedCoinsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
	finishedStreamsDistr := k.getDistributedCoinsFromIterator(ctx, k.FinishedStreamsIterator(ctx))
	return activeStreamsDistr.Add(finishedStreamsDistr...)
}
