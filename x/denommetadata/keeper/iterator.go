package keeper

import (
	"encoding/json"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	db "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// iterator returns an iterator over all denom metadata in the {prefix} space of state.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// DenomMetadataIterator returns the iterator for all denom metadata.
func (k Keeper) DenomMetadataIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixIdDenomMetadata)
}

// getDenomMetadataFromIterator iterates over everything, until it reaches the end. Return all metadata iterated over.
func (k Keeper) getDenomMetadataFromIterator(ctx sdk.Context, iterator db.Iterator) []types.DenomMetadata {
	denomMetadatas := []types.DenomMetadata{}
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		denomMetadataIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &denomMetadataIDs)
		if err != nil {
			panic(err)
		}
		for _, denomID := range denomMetadataIDs {
			denomMetadata, err := k.GetDenomMetadataByID(ctx, denomID)
			if err != nil {
				panic(err)
			}
			denomMetadatas = append(denomMetadatas, *denomMetadata)
		}
	}
	return denomMetadatas
}
