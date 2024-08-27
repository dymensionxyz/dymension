package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// NewQueryServer creates a new querier for iro clients.
func NewQueryServer(k Keeper) types.QueryServer {
	return Keeper(k)
}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

// QueryClaimed implements types.QueryServer.
func (k Keeper) QueryClaimed(context.Context, *types.QueryClaimedRequest) (*types.QueryClaimedResponse, error) {
	panic("unimplemented")
}

// QueryCost implements types.QueryServer.
func (k Keeper) QueryCost(context.Context, *types.QueryCostRequest) (*types.QueryCostResponse, error) {
	panic("unimplemented")
}

// QueryPlan implements types.QueryServer.
func (k Keeper) QueryPlan(context.Context, *types.QueryPlanRequest) (*types.QueryPlanResponse, error) {
	panic("unimplemented")
}

// QueryPlanByRollapp implements types.QueryServer.
func (k Keeper) QueryPlanByRollapp(context.Context, *types.QueryPlanByRollappRequest) (*types.QueryPlanByRollappResponse, error) {
	panic("unimplemented")
}

// QueryPlans implements types.QueryServer.
func (k Keeper) QueryPlans(context.Context, *types.QueryPlansRequest) (*types.QueryPlansResponse, error) {
	panic("unimplemented")
}

// QueryPrice implements types.QueryServer.
func (k Keeper) QueryPrice(context.Context, *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	panic("unimplemented")
}
