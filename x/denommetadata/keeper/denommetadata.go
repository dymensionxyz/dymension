package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/gogo/protobuf/proto"
)

// GetDenomMetadataByID returns denommetadata from denommetadata ID.
func (k Keeper) GetDenomMetadataByID(ctx sdk.Context, denomMetadataID uint64) (*types.DenomMetadata, error) {
	denommetadata := types.DenomMetadata{}
	store := ctx.KVStore(k.storeKey)
	denomMetadataKey := denomMetadataStoreKey(denomMetadataID)
	if !store.Has(denomMetadataKey) {
		return nil, fmt.Errorf("DenomMetadata with ID %d does not exist", denomMetadataID)
	}
	bz := store.Get(denomMetadataKey)
	if err := proto.Unmarshal(bz, &denommetadata); err != nil {
		return nil, err
	}
	return &denommetadata, nil
}

// GetStreams returns upcoming, active, and finished streams.
func (k Keeper) GetAllDenomMetadata(ctx sdk.Context) []types.DenomMetadata {
	streams := k.getDenomMetadataFromIterator(ctx, k.DenomMetadataIterator(ctx))
	// Assuming streams is your []Stream slice
	sort.Slice(streams, func(i, j int) bool {
		return streams[i].Id < streams[j].Id
	})
	return streams
}
