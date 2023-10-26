package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/dymensionxyz/dymension/x/streamer/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the streamer module keeper providing gRPC method handlers.
type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// ModuleToDistributeCoins returns coins that are going to be distributed.
func (q Querier) ModuleToDistributeCoins(goCtx context.Context, _ *types.ModuleToDistributeCoinsRequest) (*types.ModuleToDistributeCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleToDistributeCoinsResponse{Coins: q.Keeper.GetModuleToDistributeCoins(ctx)}, nil
}

// StreamByID takes a streamID and returns its respective stream.
func (q Querier) StreamByID(goCtx context.Context, req *types.StreamByIDRequest) (*types.StreamByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stream, err := q.Keeper.GetStreamByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.StreamByIDResponse{Stream: stream}, nil
}

// Streams returns all upcoming and active streams.
func (q Querier) Streams(goCtx context.Context, req *types.StreamsRequest) (*types.StreamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, streams, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixStreams, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.StreamsResponse{Data: streams, Pagination: pageRes}, nil
}

// ActiveStreams returns all active streams.
func (q Querier) ActiveStreams(goCtx context.Context, req *types.ActiveStreamsRequest) (*types.ActiveStreamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageRes, streams, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixActiveStreams, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveStreamsResponse{Data: streams, Pagination: pageRes}, nil
}

// ActiveStreamsPerDenom returns all active streams for the specified denom.
func (q Querier) ActiveStreamsPerDenom(goCtx context.Context, req *types.ActiveStreamsPerDenomRequest) (*types.ActiveStreamsPerDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pageRes, streams, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixActiveStreams, req.Denom, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.ActiveStreamsPerDenomResponse{Data: streams, Pagination: pageRes}, nil
}

// UpcomingStreams returns all upcoming streams.
func (q Querier) UpcomingStreams(goCtx context.Context, req *types.UpcomingStreamsRequest) (*types.UpcomingStreamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pageRes, streams, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixUpcomingStreams, "", req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingStreamsResponse{Data: streams, Pagination: pageRes}, nil
}

// UpcomingStreamsPerDenom returns all upcoming streams for the specified denom.
func (q Querier) UpcomingStreamsPerDenom(goCtx context.Context, req *types.UpcomingStreamsPerDenomRequest) (*types.UpcomingStreamsPerDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	pageRes, streams, err := q.filterByPrefixAndDenom(ctx, types.KeyPrefixUpcomingStreams, req.Denom, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.UpcomingStreamsPerDenomResponse{UpcomingStreams: streams, Pagination: pageRes}, nil
}

// getStreamFromIDJsonBytes returns streams from the json bytes of streamIDs.
func (q Querier) getStreamFromIDJsonBytes(ctx sdk.Context, refValue []byte) ([]types.Stream, error) {
	streams := []types.Stream{}
	streamIDs := []uint64{}

	err := json.Unmarshal(refValue, &streamIDs)
	if err != nil {
		return streams, err
	}

	for _, streamID := range streamIDs {
		stream, err := q.Keeper.GetStreamByID(ctx, streamID)
		if err != nil {
			return []types.Stream{}, err
		}

		streams = append(streams, *stream)
	}

	return streams, nil
}

// filterByPrefixAndDenom filters streams based on a given key prefix and denom
func (q Querier) filterByPrefixAndDenom(ctx sdk.Context, prefixType []byte, denom string, pagination *query.PageRequest) (*query.PageResponse, []types.Stream, error) {
	streams := []types.Stream{}
	store := ctx.KVStore(q.Keeper.storeKey)
	valStore := prefix.NewStore(store, prefixType)

	pageRes, err := query.FilteredPaginate(valStore, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// this may return multiple streams at once if two streams start at the same time.
		// for now this is treated as an edge case that is not of importance
		newStreams, err := q.getStreamFromIDJsonBytes(ctx, value)
		if err != nil {
			return false, err
		}
		if accumulate {
			streams = append(streams, newStreams...)
		}
		return true, nil
	})
	return pageRes, streams, err
}
