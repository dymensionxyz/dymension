package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// AddStreamRefByKey appends the provided stream ID into an array associated with the provided key.
func (k Keeper) AddStreamRefByKey(ctx sdk.Context, key []byte, streamID uint64) error {
	return k.addStreamRefByKey(ctx, key, streamID)
}

// DeleteStreamRefByKey removes the provided stream ID from an array associated with the provided key.
func (k Keeper) DeleteStreamRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.deleteStreamRefByKey(ctx, key, guageID)
}

// GetStreamRefs returns the stream IDs specified by the provided key.
func (k Keeper) GetStreamRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getStreamRefs(ctx, key)
}

// MoveUpcomingStreamToActiveStream moves a stream that has reached it's start time from an upcoming to an active status.
func (k Keeper) MoveUpcomingStreamToActiveStream(ctx sdk.Context, stream types.Stream) error {
	return k.moveUpcomingStreamToActiveStream(ctx, stream)
}

// MoveActiveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
func (k Keeper) MoveActiveStreamToFinishedStream(ctx sdk.Context, stream types.Stream) error {
	return k.moveActiveStreamToFinishedStream(ctx, stream)
}
