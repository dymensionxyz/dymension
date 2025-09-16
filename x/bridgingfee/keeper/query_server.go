package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// FeeHook returns a fee hook by ID
func (k queryServer) FeeHook(ctx context.Context, req *types.QueryFeeHookRequest) (*types.QueryFeeHookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	hook, err := k.feeHooks.Get(ctx, req.Id.GetInternalId())
	if err != nil {
		return nil, errorsmod.Wrap(err, "get fee hook")
	}

	return &types.QueryFeeHookResponse{FeeHook: hook}, nil
}

// FeeHooks returns all fee hooks
func (k queryServer) FeeHooks(ctx context.Context, req *types.QueryFeeHooksRequest) (*types.QueryFeeHooksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var hooks []types.HLFeeHook
	err := k.feeHooks.Walk(ctx, nil, func(key uint64, value types.HLFeeHook) (stop bool, err error) {
		hooks = append(hooks, value)
		return false, nil
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "walk fee hooks")
	}

	return &types.QueryFeeHooksResponse{FeeHooks: hooks}, nil
}

// AggregationHook returns an aggregation hook by ID
func (k queryServer) AggregationHook(ctx context.Context, req *types.QueryAggregationHookRequest) (*types.QueryAggregationHookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	hook, err := k.aggregationHooks.Get(ctx, req.Id.GetInternalId())
	if err != nil {
		return nil, errorsmod.Wrap(err, "get aggregation hook")
	}

	return &types.QueryAggregationHookResponse{AggregationHook: hook}, nil
}

// AggregationHooks returns all aggregation hooks
func (k queryServer) AggregationHooks(ctx context.Context, req *types.QueryAggregationHooksRequest) (*types.QueryAggregationHooksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var hooks []types.AggregationHook
	err := k.aggregationHooks.Walk(ctx, nil, func(key uint64, value types.AggregationHook) (stop bool, err error) {
		hooks = append(hooks, value)
		return false, nil
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "walk aggregation hooks")
	}

	return &types.QueryAggregationHooksResponse{AggregationHooks: hooks}, nil
}