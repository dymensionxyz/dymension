package keeper

import (
	"fmt"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// UpdateStreamAtEpochStart updates the stream for a new epoch: estimates coins that streamer will
// distribute during this epoch and updates a sponsored distribution if needed.
func (k Keeper) UpdateStreamAtEpochStart(ctx sdk.Context, stream types.Stream) (types.Stream, error) {
	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	remainEpochs := stream.NumEpochsPaidOver - stream.FilledEpochs
	epochCoins := remainCoins.QuoInt(math.NewIntFromUint64(remainEpochs))

	// If the stream uses a sponsorship plan, query it and update stream distr info. The distribution
	// might be empty and this is a valid scenario. In that case, we'll just skip without filling the epoch.
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

	return stream, nil
}

// UpdateStreamAtEpochEnd updates the stream at the end of the epoch: increases the filled epoch number
// and makes the stream finished if needed.
func (k Keeper) UpdateStreamAtEpochEnd(ctx sdk.Context, stream types.Stream) (types.Stream, error) {
	// Don't fill streams in which there's nothing to fill. This might happen when using sponsored streams.
	if !stream.DistributeTo.TotalWeight.IsZero() {
		stream.FilledEpochs += 1
	}

	// Check if stream has completed its distribution. This is a post factum check.
	if stream.FilledEpochs >= stream.NumEpochsPaidOver {
		err := k.moveActiveStreamToFinishedStream(ctx, stream)
		if err != nil {
			return types.Stream{}, fmt.Errorf("move active stream to finished stream: %w", err)
		}
	}

	return stream, nil
}

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

// GetStreams returns upcoming, active, and finished streams.
func (k Keeper) GetStreams(ctx sdk.Context) []types.Stream {
	streams := k.getStreamsFromIterator(ctx, k.StreamsIterator(ctx))
	// Assuming streams is your []Stream slice
	sort.Slice(streams, func(i, j int) bool {
		return streams[i].Id < streams[j].Id
	})
	return streams
}

// GetNotFinishedStreams returns both upcoming and active streams.
func (k Keeper) GetNotFinishedStreams(ctx sdk.Context) []types.Stream {
	return append(k.GetActiveStreams(ctx), k.GetUpcomingStreams(ctx)...)
}

// GetActiveStreams returns active streams.
func (k Keeper) GetActiveStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
}

// GetActiveStreamsForEpoch returns active streams with the specified epoch identifier.
func (k Keeper) GetActiveStreamsForEpoch(ctx sdk.Context, epochIdentifier string) []types.Stream {
	streams := k.getStreamsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
	activeStreams := make([]types.Stream, 0)
	for _, stream := range streams {
		if stream.DistrEpochIdentifier == epochIdentifier {
			activeStreams = append(activeStreams, stream)
		}
	}
	return activeStreams
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

// moveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
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
