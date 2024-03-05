package keeper

import (
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

// streamStoreKey returns the combined byte array (store key) of the provided stream ID's key prefix and the ID itself.
func denomMetadataStoreKey(ID uint64) []byte {
	return combineKeys(types.KeyPrefixPeriodDenomMetadata, sdk.Uint64ToBigEndian(ID))
}

// setStream set the stream inside store.
func (k Keeper) setDenomMetadata(ctx sdk.Context, denomMetadata *types.DenomMetadata) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(denomMetadata)
	if err != nil {
		return err
	}
	store.Set(denomMetadataStoreKey(denomMetadata.Id), bz)
	return nil
}
