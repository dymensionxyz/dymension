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

// streamDenomStoreKey returns the combined byte array (store key) of the provided stream denom key prefix and the denom itself.
func streamDenomStoreKey(denom string) []byte {
	return combineKeys(types.KeyPrefixStreamsByDenom, []byte(denom))
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

// getAllStreamIDsByDenom returns all active stream-IDs associated with lockups of the provided denom.
func (k Keeper) getAllStreamIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getStreamRefs(ctx, streamDenomStoreKey(denom))
}

// deleteStreamIDForDenom deletes the provided ID from the list of stream ID's associated with the provided denom.
func (k Keeper) deleteStreamIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.deleteStreamRefByKey(ctx, streamDenomStoreKey(denom), ID)
}

// addStreamIDForDenom adds the provided ID to the list of stream ID's associated with the provided denom.
func (k Keeper) addStreamIDForDenom(ctx sdk.Context, ID uint64, denom string) error {
	return k.addStreamRefByKey(ctx, streamDenomStoreKey(denom), ID)
}
