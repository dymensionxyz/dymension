package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastStreamID returns the last used stream ID.
func (k Keeper) GetLastStreamID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastStreamID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastStreamID sets the last used stream ID to the provided ID.
func (k Keeper) SetLastStreamID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastStreamID, sdk.Uint64ToBigEndian(ID))
}

// CreateStreamRefKeys takes combinedKey (the keyPrefix for upcoming, active, or finished streams combined with stream start time) and adds a reference to the respective stream ID.
// If stream is active or upcoming, creates reference between the denom and stream ID.
// Used to consolidate codepaths for InitGenesis and CreateStream.
func (k Keeper) CreateStreamRefKeys(ctx sdk.Context, stream *types.Stream, combinedKeys []byte) error {
	if err := k.addStreamRefByKey(ctx, combinedKeys, stream.Id); err != nil {
		return err
	}

	return nil
}

// SetStreamWithRefKey takes a single stream and assigns a key.
// Takes combinedKey (the keyPrefix for upcoming, active, or finished streams combined with stream start time) and adds a reference to the respective stream ID.
func (k Keeper) SetStreamWithRefKey(ctx sdk.Context, stream *types.Stream) error {
	err := k.setStream(ctx, stream)
	if err != nil {
		return err
	}

	curTime := ctx.BlockTime()
	timeKey := getTimeKey(stream.StartTime)

	if stream.IsUpcomingStream(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixUpcomingStreams, timeKey)
		return k.CreateStreamRefKeys(ctx, stream, combinedKeys)
	} else if stream.IsActiveStream(curTime) {
		combinedKeys := combineKeys(types.KeyPrefixActiveStreams, timeKey)
		return k.CreateStreamRefKeys(ctx, stream, combinedKeys)
	} else {
		combinedKeys := combineKeys(types.KeyPrefixFinishedStreams, timeKey)
		return k.CreateStreamRefKeys(ctx, stream, combinedKeys)
	}
}

// streamStoreKey returns the combined byte array (store key) of the provided stream ID's key prefix and the ID itself.
func streamStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodStream, sdk.Uint64ToBigEndian(ID))
}

// setStream set the stream inside store.
func (k Keeper) setStream(ctx sdk.Context, stream *types.Stream) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(stream)
	if err != nil {
		return err
	}
	store.Set(streamStoreKey(stream.Id), bz)
	return nil
}

// getStreamRefs returns the stream IDs specified by the provided key.
func (k Keeper) getStreamRefs(ctx sdk.Context, key []byte) []uint64 {
	store := ctx.KVStore(k.storeKey)
	streamIDs := []uint64{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &streamIDs)
		if err != nil {
			panic(err)
		}
	}
	return streamIDs
}

// addStreamRefByKey appends the provided stream ID into an array associated with the provided key.
func (k Keeper) addStreamRefByKey(ctx sdk.Context, key []byte, streamID uint64) error {
	store := ctx.KVStore(k.storeKey)
	streamIDs := k.getStreamRefs(ctx, key)
	if findIndex(streamIDs, streamID) > -1 {
		return fmt.Errorf("stream with same ID exist: %d", streamID)
	}
	streamIDs = append(streamIDs, streamID)
	bz, err := json.Marshal(streamIDs)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// deleteStreamRefByKey removes the provided stream ID from an array associated with the provided key.
func (k Keeper) deleteStreamRefByKey(ctx sdk.Context, key []byte, streamID uint64) error {
	store := ctx.KVStore(k.storeKey)
	streamIDs := k.getStreamRefs(ctx, key)
	streamIDs, index := removeValue(streamIDs, streamID)
	if index < 0 {
		return fmt.Errorf("specific stream with ID %d not found by reference %s", streamID, key)
	}
	if len(streamIDs) == 0 {
		store.Delete(key)
	} else {
		bz, err := json.Marshal(streamIDs)
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	return nil
}
