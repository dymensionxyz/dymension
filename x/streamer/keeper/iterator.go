package keeper

import (
	"encoding/json"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	db "github.com/tendermint/tm-db"

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

// getStreamsFromIterator iterates over everything in a stream's iterator, until it reaches the end. Return all streams iterated over.
func (k Keeper) getStreamsFromIterator(ctx sdk.Context, iterator db.Iterator) []types.Stream {
	streams := []types.Stream{}
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		streamIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &streamIDs)
		if err != nil {
			panic(err)
		}
		for _, streamID := range streamIDs {
			stream, err := k.GetStreamByID(ctx, streamID)
			if err != nil {
				panic(err)
			}
			streams = append(streams, *stream)
		}
	}
	return streams
}
