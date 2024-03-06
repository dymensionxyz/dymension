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

// SetDenomMetadataWithRefKey sets denommedatata into the store, including extra keys used to check uniqueness
func (k Keeper) SetDenomMetadataWithRefKey(ctx sdk.Context, denomMetadata *types.DenomMetadata) error {
	err := k.setDenomMetadata(ctx, denomMetadata)
	if err != nil {
		return err
	}

	if err := k.addDenomMetadataRefByKey(ctx, combineKeys(types.KeyPrefixBaseDenomMetadata, []byte(denomMetadata.TokenMetadata.Base)), denomMetadata.Id); err != nil {
		return err
	}
	/*if err := k.addDenomMetadataRefByKey(ctx, combineKeys(types.KeyPrefixDisplayDenomMetadata, []byte(denomMetadata.TokenMetadata.Display)), denomMetadata.Id); err != nil {
		return err
	}
	if err := k.addDenomMetadataRefByKey(ctx, combineKeys(types.KeyPrefixSymbolDenomMetadata, []byte(denomMetadata.TokenMetadata.Symbol)), denomMetadata.Id); err != nil {
		return err
	}*/
	return nil

}

// SetLastDenomMetadataID sets the last used DenomMetadata ID to the provided ID.
func (k Keeper) SetLastDenomMetadataID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastDenomMetadataID, sdk.Uint64ToBigEndian(ID))
}

// denomMetadataStoreKey returns the combined byte array (store key) of the provided denommetadata ID's key prefix and the ID itself.
func denomMetadataStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixIdDenomMetadata, sdk.Uint64ToBigEndian(ID))
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
	denomIDs := []uint64{}
	if store.Has(key) {
		bz := store.Get(key)
		err := json.Unmarshal(bz, &denomIDs)
		if err != nil {
			panic(err)
		}
	}
	return denomIDs
}

// addDenomMetadataRefByKey appends the denommetadata denom ID into an array associated with the provided key.
func (k Keeper) addDenomMetadataRefByKey(ctx sdk.Context, key []byte, denomID uint64) error {
	store := ctx.KVStore(k.storeKey)
	denomIDs := k.getDenomMetadataRefs(ctx, key)
	if findIndex(denomIDs, denomID) > -1 {
		return fmt.Errorf("denom metadata with same ID exist: %d", denomID)
	}
	denomIDs = append(denomIDs, denomID)
	bz, err := json.Marshal(denomIDs)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}
