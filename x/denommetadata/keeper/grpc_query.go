package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	denommetadatas := q.Keeper.GetAllDenomMetadata(ctx)

	return &types.AllDenomMetadataResponse{Data: denommetadatas}, nil
}
