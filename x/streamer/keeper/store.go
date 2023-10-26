package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/dymensionxyz/dymension/x/streamer/types"

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

// streamStoreKey returns the combined byte array (store key) of the provided stream ID's key prefix and the ID itself.
func streamStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodStream, sdk.Uint64ToBigEndian(ID))
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
