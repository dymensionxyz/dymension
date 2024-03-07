package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/gogo/protobuf/proto"
)

// GetDenomMetadataByID returns denommetadata from denommetadata ID.
func (k Keeper) GetDenomMetadataByID(ctx sdk.Context, denomMetadataID uint64) (*types.DenomMetadata, error) {
	denommetadata := types.DenomMetadata{}
	store := ctx.KVStore(k.storeKey)
	denomMetadataKey := denomMetadataStoreIdKey(denomMetadataID)
	if !store.Has(denomMetadataKey) {
		return nil, fmt.Errorf("DenomMetadata with ID %d does not exist", denomMetadataID)
	}
	bz := store.Get(denomMetadataKey)
	if err := proto.Unmarshal(bz, &denommetadata); err != nil {
		return nil, err
	}
	return &denommetadata, nil
}

// GetDenomMetadataByBaseDenom returns denommetadata from denommetadata base param.
func (k Keeper) GetDenomMetadataByBaseDenom(ctx sdk.Context, base string) (*types.DenomMetadata, error) {
	store := ctx.KVStore(k.storeKey)
	denomMetadataBaseKey := denomMetadataStoreBaseKey(base)
	if !store.Has(denomMetadataBaseKey) {
		return nil, fmt.Errorf("DenomMetadata with base %s does not exist", base)
	}
	key := store.Get(denomMetadataBaseKey)
	denommetadata := types.DenomMetadata{}
	if !store.Has(key) {
		return nil, fmt.Errorf("DenomMetadata with ID %d does not exist", key)
	}
	bz := store.Get(key)
	if err := proto.Unmarshal(bz, &denommetadata); err != nil {
		return nil, err
	}
	return &denommetadata, nil
}

// GetDenomMetadataByBaseDenom returns denommetadata from denommetadata base param.
func (k Keeper) GetDenomMetadataByDisplayDenom(ctx sdk.Context, display string) (*types.DenomMetadata, error) {
	store := ctx.KVStore(k.storeKey)
	denomMetadataBaseKey := denomMetadataStoreDisplayKey(display)
	if !store.Has(denomMetadataBaseKey) {
		return nil, fmt.Errorf("DenomMetadata with base %s does not exist", display)
	}
	key := store.Get(denomMetadataBaseKey)
	denommetadata := types.DenomMetadata{}
	if !store.Has(key) {
		return nil, fmt.Errorf("DenomMetadata with ID %d does not exist", key)
	}
	bz := store.Get(key)
	if err := proto.Unmarshal(bz, &denommetadata); err != nil {
		return nil, err
	}
	return &denommetadata, nil
}

// GetDenomMetadataByBaseDenom returns denommetadata from denommetadata base param.
func (k Keeper) GetDenomMetadataBySymbolDenom(ctx sdk.Context, symbol string) (*types.DenomMetadata, error) {
	store := ctx.KVStore(k.storeKey)
	denomMetadataBaseKey := denomMetadataStoreSymbolKey(symbol)
	if !store.Has(denomMetadataBaseKey) {
		return nil, fmt.Errorf("DenomMetadata with base %s does not exist", symbol)
	}
	key := store.Get(denomMetadataBaseKey)
	denommetadata := types.DenomMetadata{}
	if !store.Has(key) {
		return nil, fmt.Errorf("DenomMetadata with ID %d does not exist", key)
	}
	bz := store.Get(key)
	if err := proto.Unmarshal(bz, &denommetadata); err != nil {
		return nil, err
	}
	return &denommetadata, nil
}

// GetAllDenomMetadata returns all registered denoms.
func (k Keeper) GetAllDenomMetadata(ctx sdk.Context) []types.DenomMetadata {

	denomMetadata := []types.DenomMetadata{}
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixIdDenomMetadata)
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.DenomMetadata
		proto.Unmarshal(iterator.Value(), &val)
		denomMetadata = append(denomMetadata, val)
	}
	return denomMetadata
}
