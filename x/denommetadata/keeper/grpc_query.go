package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the denometadata module keeper providing gRPC method handlers.
type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// DenomMetadataByID takes a denomMetadataID and returns its respective metadata.
func (q Querier) DenomMetadataByID(goCtx context.Context, req *types.DenomMetadataByIDRequest) (*types.DenomMetadataByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	denonmetadata, err := q.Keeper.GetDenomMetadataByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.DenomMetadataByIDResponse{Metadata: denonmetadata}, nil
}

// DenomMetadataByBaseDenom returns denom metadata by base denom.
func (q Querier) DenomMetadataByBaseDenom(goCtx context.Context, req *types.DenomMetadataByBaseDenomRequest) (*types.DenomMetadataByBaseDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	denonmetadata, err := q.Keeper.GetDenomMetadataByBaseDenom(ctx, req.BaseDenom)
	if err != nil {
		return nil, err
	}
	return &types.DenomMetadataByBaseDenomResponse{Metadata: denonmetadata}, nil
}

// DenomMetadataByDisplayDenom returns denom metadata by display denom.
func (q Querier) DenomMetadataByDisplayDenom(goCtx context.Context, req *types.DenomMetadataByDisplayDenomRequest) (*types.DenomMetadataByDisplayDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	denonmetadata, err := q.Keeper.GetDenomMetadataByBaseDenom(ctx, req.DisplayDenom)
	if err != nil {
		return nil, err
	}
	return &types.DenomMetadataByDisplayDenomResponse{Metadata: denonmetadata}, nil
}

// DenomMetadataBySymbolDenom returns denom metadata by symbol denom.
func (q Querier) DenomMetadataBySymbolDenom(goCtx context.Context, req *types.DenomMetadataBySymbolDenomRequest) (*types.DenomMetadataBySymbolDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	denonmetadata, err := q.Keeper.GetDenomMetadataByBaseDenom(ctx, req.SymbolDenom)
	if err != nil {
		return nil, err
	}
	return &types.DenomMetadataBySymbolDenomResponse{Metadata: denonmetadata}, nil
}

// AllDenomMetadata returns all denom metadata registered.
func (q Querier) AllDenomMetadata(goCtx context.Context, req *types.AllDenomMetadataRequest) (*types.AllDenomMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, denommetadatas, err := q.filterByPrefixAndDenom(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.AllDenomMetadataResponse{Data: denommetadatas, Pagination: pageRes}, nil
}

// filterByPrefixAndDenom filters streams based on a given key prefix and denom
func (q Querier) filterByPrefixAndDenom(ctx sdk.Context, pagination *query.PageRequest) (*query.PageResponse, []types.DenomMetadata, error) {
	denommetadatas := []types.DenomMetadata{}
	store := ctx.KVStore(q.Keeper.storeKey)

	pageRes, err := query.FilteredPaginate(store, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// this may return multiple streams at once if two streams start at the same time.
		// for now this is treated as an edge case that is not of importance
		newdenoms, err := q.getDenomMetadataFromIDJsonBytes(ctx, value)
		if err != nil {
			return false, err
		}
		if accumulate {
			denommetadatas = append(denommetadatas, newdenoms...)
		}
		return true, nil
	})
	return pageRes, denommetadatas, err
}

// getStreamFromIDJsonBytes returns streams from the json bytes of streamIDs.
func (q Querier) getDenomMetadataFromIDJsonBytes(ctx sdk.Context, refValue []byte) ([]types.DenomMetadata, error) {
	denommetadatas := []types.DenomMetadata{}
	denommetadataIDs := []uint64{}

	err := json.Unmarshal(refValue, &denommetadataIDs)
	if err != nil {
		return denommetadatas, err
	}

	for _, denomID := range denommetadataIDs {
		denommetadata, err := q.Keeper.GetDenomMetadataByID(ctx, denomID)
		if err != nil {
			return []types.DenomMetadata{}, err
		}

		denommetadatas = append(denommetadatas, *denommetadata)
	}

	return denommetadatas, nil
}
