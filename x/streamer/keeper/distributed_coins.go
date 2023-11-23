package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/streamer/types"
	db "github.com/tendermint/tm-db"
)

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
