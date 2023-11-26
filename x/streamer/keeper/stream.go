package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetStreamByID returns stream from stream ID.
func (k Keeper) GetStreamByID(ctx sdk.Context, streamID uint64) (*types.Stream, error) {
	stream := types.Stream{}
	store := ctx.KVStore(k.storeKey)
	streamKey := streamStoreKey(streamID)
	if !store.Has(streamKey) {
		return nil, fmt.Errorf("stream with ID %d does not exist", streamID)
	}
	bz := store.Get(streamKey)
	if err := proto.Unmarshal(bz, &stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

// GetStreamFromIDs returns multiple streams from a streamIDs array.
func (k Keeper) GetStreamFromIDs(ctx sdk.Context, streamIDs []uint64) ([]types.Stream, error) {
	streams := []types.Stream{}
	for _, streamID := range streamIDs {
		stream, err := k.GetStreamByID(ctx, streamID)
		if err != nil {
			return []types.Stream{}, err
		}
		streams = append(streams, *stream)
	}
	return streams, nil
}

// GetStreams returns upcoming, active, and finished streams.
func (k Keeper) GetStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.StreamsIterator(ctx))
}

// GetNotFinishedStreams returns both upcoming and active streams.
func (k Keeper) GetNotFinishedStreams(ctx sdk.Context) []types.Stream {
	return append(k.GetActiveStreams(ctx), k.GetUpcomingStreams(ctx)...)
}

// GetActiveStreams returns active streams.
func (k Keeper) GetActiveStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
}

// GetUpcomingStreams returns upcoming streams.
func (k Keeper) GetUpcomingStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.UpcomingStreamsIterator(ctx))
}

// GetFinishedStreams returns finished streams.
func (k Keeper) GetFinishedStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.FinishedStreamsIterator(ctx))
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

// moveUpcomingStreamToFinishedStream moves a stream that is still upcoming to a finished status.
func (k Keeper) moveUpcomingStreamToFinishedStream(ctx sdk.Context, stream types.Stream) error {
	return k.moveStreamToFinishedStream(ctx, stream, types.KeyPrefixUpcomingStreams)
}

// moveActiveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveStreamToFinishedStream(ctx sdk.Context, stream types.Stream) error {
	return k.moveStreamToFinishedStream(ctx, stream, types.KeyPrefixActiveStreams)
}

// moveActiveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
func (k Keeper) moveStreamToFinishedStream(ctx sdk.Context, stream types.Stream, prefixKey []byte) error {
	timeKey := getTimeKey(stream.StartTime)
	if err := k.deleteStreamRefByKey(ctx, combineKeys(prefixKey, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.addStreamRefByKey(ctx, combineKeys(types.KeyPrefixFinishedStreams, timeKey), stream.Id); err != nil {
		return err
	}
	return nil
}
