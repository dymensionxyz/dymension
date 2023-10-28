package keeper

import (
	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// iterator returns an iterator over all streams in the {prefix} space of state.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// StreamsIterator returns the iterator for all streams.
func (k Keeper) StreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixStreams)
}

// UpcomingStreamsIterator returns the iterator for all upcoming streams.
func (k Keeper) UpcomingStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixUpcomingStreams)
}

// ActiveStreamsIterator returns the iterator for all active streams.
func (k Keeper) ActiveStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixActiveStreams)
}

// FinishedStreamsIterator returns the iterator for all finished streams.
func (k Keeper) FinishedStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixFinishedStreams)
}
