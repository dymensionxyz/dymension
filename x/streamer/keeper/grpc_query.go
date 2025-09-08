package keeper

import (
	"context"
	"encoding/json"
	"slices"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
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

// Params returns the total set of streamer parameters.
func (q Querier) Params(goCtx context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.GetParams(ctx)
	return &types.ParamsResponse{Params: params}, nil
}

// ModuleToDistributeCoins returns coins that are going to be distributed.
func (q Querier) ModuleToDistributeCoins(goCtx context.Context, _ *types.ModuleToDistributeCoinsRequest) (*types.ModuleToDistributeCoinsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleToDistributeCoinsResponse{Coins: q.GetModuleToDistributeCoins(ctx)}, nil
}

// StreamByID takes a streamID and returns its respective stream.
func (q Querier) StreamByID(goCtx context.Context, req *types.StreamByIDRequest) (*types.StreamByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stream, err := q.GetStreamByID(ctx, req.Id)
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

// PumpPressure for every RA_i is calculated as RA_i / ∑^N RA_j
func (q Querier) PumpPressure(goCtx context.Context, req *types.PumpPressureRequest) (*types.PumpPressureResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	d, err := q.sk.GetDistribution(ctx)
	if err != nil {
		return nil, err
	}

	totalPressure := q.TotalPumpBudget(ctx)
	pressure := q.Keeper.TopRollapps(ctx, d.Gauges, totalPressure, nil)

	return &types.PumpPressureResponse{
		Pressure:   pressure,
		Pagination: nil, // TODO: pagination?
	}, nil
}

// PumpPressureByRollapp for RA_i is calculated as RA_i / ∑^N RA_j
func (q Querier) PumpPressureByRollapp(goCtx context.Context, req *types.PumpPressureByRollappRequest) (*types.PumpPressureByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	d, err := q.sk.GetDistribution(ctx)
	if err != nil {
		return nil, err
	}

	totalPressure := q.TotalPumpBudget(ctx)
	pressure := q.Keeper.TopRollapps(ctx, d.Gauges, totalPressure, nil)
	idx := slices.IndexFunc(pressure, func(p types.PumpPressure) bool {
		return p.RollappId == req.RollappId
	})
	if idx < 0 {
		return nil, status.Error(codes.NotFound, "rollapp don't have any pressure")
	}

	return &types.PumpPressureByRollappResponse{
		Pressure: pressure[idx],
	}, nil
}

// PumpPressureByUser for U and for each RA_u_i, which got a vote from U,
// is calculated as RA_u_i / ∑^N RA_j
func (q Querier) PumpPressureByUser(goCtx context.Context, req *types.PumpPressureByUserRequest) (*types.PumpPressureByUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	userAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	vote, err := q.sk.GetVote(ctx, userAddr)
	if err != nil {
		return nil, err
	}
	d, err := q.sk.GetDistribution(ctx)
	if err != nil {
		return nil, err
	}

	totalWeight := math.ZeroInt()
	for _, gauge := range d.Gauges {
		totalWeight = totalWeight.Add(gauge.Power)
	}

	// We need to see how the user contributes to the total distribution.
	// Use total voting power with user's weights to get PumpPressure.
	totalPressure := q.TotalPumpBudget(ctx)
	pressure := q.Keeper.PumpPressure(ctx, vote.ToDistribution().Gauges, totalPressure, totalWeight)

	return &types.PumpPressureByUserResponse{
		Pressure:   pressure,
		Pagination: nil, // TODO: pagination?
	}, nil
}

// PumpPressureByUserByRollapp for U and for RA_u_i, which got a vote from U,
// is calculated as RA_u_i / ∑^N RA_j
func (q Querier) PumpPressureByUserByRollapp(goCtx context.Context, req *types.PumpPressureByUserByRollappRequest) (*types.PumpPressureByUserByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	userAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	vote, err := q.sk.GetVote(ctx, userAddr)
	if err != nil {
		return nil, err
	}
	d, err := q.sk.GetDistribution(ctx)
	if err != nil {
		return nil, err
	}

	totalWeight := math.ZeroInt()
	for _, gauge := range d.Gauges {
		totalWeight = totalWeight.Add(gauge.Power)
	}

	// We need to see how the user contributes to the total distribution.
	// Use total voting power with user's weights to get PumpPressure.
	totalPressure := q.TotalPumpBudget(ctx)
	pressure := q.Keeper.PumpPressure(ctx, vote.ToDistribution().Gauges, totalPressure, totalWeight)
	idx := slices.IndexFunc(pressure, func(p types.PumpPressure) bool {
		return p.RollappId == req.RollappId
	})
	if idx < 0 {
		return nil, status.Error(codes.NotFound, "rollapp don't have any pressure")
	}

	return &types.PumpPressureByUserByRollappResponse{
		Pressure: pressure[idx],
	}, nil
}

// getStreamFromIDJsonBytes returns streams from the json bytes of streamIDs.
func (k Keeper) getStreamFromIDJsonBytes(ctx sdk.Context, refValue []byte) ([]types.Stream, error) {
	streams := []types.Stream{}
	streamIDs := []uint64{}

	err := json.Unmarshal(refValue, &streamIDs)
	if err != nil {
		return streams, err
	}

	for _, streamID := range streamIDs {
		stream, err := k.GetStreamByID(ctx, streamID)
		if err != nil {
			return []types.Stream{}, err
		}

		streams = append(streams, *stream)
	}

	return streams, nil
}

// FIXME: denom not used
// filterByPrefixAndDenom filters streams based on a given key prefix
func (k Keeper) filterByPrefixAndDenom(ctx sdk.Context, prefixType []byte, _ string, pagination *query.PageRequest) (*query.PageResponse, []types.Stream, error) {
	streams := []types.Stream{}
	store := ctx.KVStore(k.storeKey)
	valStore := prefix.NewStore(store, prefixType)

	pageRes, err := query.FilteredPaginate(valStore, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// this may return multiple streams at once if two streams start at the same time.
		// for now this is treated as an edge case that is not of importance
		newStreams, err := k.getStreamFromIDJsonBytes(ctx, value)
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
