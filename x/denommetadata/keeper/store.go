package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetLastDenomMetadataID returns the last used DenomMetadata ID.
func (k Keeper) GetLastDenomMetadataID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastDenomMetadataID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastDenomMetadataID sets the last used DenomMetadata ID to the provided ID.
func (k Keeper) SetLastDenomMetadataID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastDenomMetadataID, sdk.Uint64ToBigEndian(ID))
}

// denomMetadataStoreKey returns the combined byte array (store key) of the provided denommetadata ID's key prefix and the ID itself.
func denomMetadataStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodDenomMetadata, sdk.Uint64ToBigEndian(ID))
}

// setDenomMetadata set the denommetadata inside store.
func (k Keeper) setDenomMetadata(ctx sdk.Context, denomMetadata *types.DenomMetadata) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(denomMetadata)
	if err != nil {
		return err
	}
	store.Set(denomMetadataStoreKey(denomMetadata.Id), bz)
	return nil
}

// getDenomMetadataRefs returns the denommetadata IDs specified by the provided key.
func (k Keeper) getDenomMetadataRefs(ctx sdk.Context, key []byte) []uint64 {
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

// addDenomMetadataRefByKey appends the denommetadata stream ID into an array associated with the provided key.
func (k Keeper) addDenomMetadataRefByKey(ctx sdk.Context, key []byte, streamID uint64) error {
	store := ctx.KVStore(k.storeKey)
	streamIDs := k.getDenomMetadataRefs(ctx, key)
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
