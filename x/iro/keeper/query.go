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
	return k
}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

// QueryClaimed implements types.QueryServer.
func (k Keeper) QueryClaimed(goCtx context.Context, req *types.QueryClaimedRequest) (*types.QueryClaimedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QueryClaimedResponse{ClaimedAmt: &plan.ClaimedAmt}, nil
}

// QueryCost implements types.QueryServer.
func (k Keeper) QueryCost(goCtx context.Context, req *types.QueryCostRequest) (*types.QueryCostResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QueryCostResponse{Cost: &plan.TotalAllocation}, nil
}

// QueryPlan implements types.QueryServer.
func (k Keeper) QueryPlan(goCtx context.Context, req *types.QueryPlanRequest) (*types.QueryPlanResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QueryPlanResponse{Plan: &plan}, nil
}

// QueryPlanByRollapp implements types.QueryServer.
func (k Keeper) QueryPlanByRollapp(goCtx context.Context, req *types.QueryPlanByRollappRequest) (*types.QueryPlanByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlanByRollapp(ctx, req.RollappId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QueryPlanByRollappResponse{Plan: &plan}, nil
}

// QueryPlans implements types.QueryServer.
func (k Keeper) QueryPlans(goCtx context.Context, req *types.QueryPlansRequest) (*types.QueryPlansResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plans := k.GetAllPlans(ctx)

	return &types.QueryPlansResponse{Plans: plans}, nil

}

// QueryPrice implements types.QueryServer.
func (k Keeper) QueryPrice(goCtx context.Context, req *types.QueryPriceRequest) (*types.QueryPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	price := plan.BondingCurve.SpotPrice(plan.SoldAmt).TruncateInt()
	coin := sdk.NewCoin(plan.TotalAllocation.Denom, price)

	// FIXME: should be Decimal price, not coin!
	return &types.QueryPriceResponse{
		Price: &coin,
	}, nil
}
