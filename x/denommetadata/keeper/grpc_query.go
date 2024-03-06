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

func (q Querier) DenomMetadataByBaseDenom(goCtx context.Context, req *types.DenomMetadataByBaseDenomRequest) (*types.DenomMetadataByBaseDenomResponse, error) {
	return &types.DenomMetadataByBaseDenomResponse{}, nil
}

func (q Querier) DenomMetadataByDisplayDenom(goCtx context.Context, req *types.DenomMetadataByDisplayDenomRequest) (*types.DenomMetadataByDisplayDenomResponse, error) {
	return &types.DenomMetadataByDisplayDenomResponse{}, nil
}

func (q Querier) DenomMetadataBySymbolDenom(goCtx context.Context, req *types.DenomMetadataBySymbolDenomRequest) (*types.DenomMetadataBySymbolDenomResponse, error) {
	return &types.DenomMetadataBySymbolDenomResponse{}, nil
}

// Streams returns all upcoming and active streams.
func (q Querier) AllDenomMetadata(goCtx context.Context, req *types.AllDenomMetadataRequest) (*types.AllDenomMetadataResponse, error) {

	return &types.AllDenomMetadataResponse{Data: nil, Pagination: nil}, nil
}
