package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
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

	var costAmt math.Int
	if req.Sell {
		costAmt = plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Sub(req.Amt))
	} else {
		costAmt = plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(req.Amt))
	}
	cost := sdk.NewCoin(plan.LiquidityDenom, costAmt)
	return &types.QueryCostResponse{Cost: &cost}, nil
}

// QueryTokensForExactInAmount implements types.QueryServer.
func (k Keeper) QueryTokensForExactInAmount(goCtx context.Context, req *types.QueryTokensForExactInAmountRequest) (*types.QueryTokensForExactInAmountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	tokensAmt, err := plan.BondingCurve.TokensForExactInAmount(plan.SoldAmt, req.Amt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tokens := sdk.NewCoin(plan.GetIRODenom(), tokensAmt)
	return &types.QueryTokensForExactInAmountResponse{Tokens: &tokens}, nil
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
	plans, pageRes, err := k.GetAllPlansPaginated(ctx, req.NonSettledOnly, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPlansResponse{Plans: plans, Pagination: pageRes}, nil
}

// QuerySpotPrice implements types.QueryServer.
func (k Keeper) QuerySpotPrice(goCtx context.Context, req *types.QuerySpotPriceRequest) (*types.QuerySpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	return &types.QuerySpotPriceResponse{
		Price: plan.SpotPrice(),
	}, nil
}

// QueryVesting implements types.QueryServer.
func (k Keeper) QueryVesting(goCtx context.Context, req *types.QueryVestingRequest) (*types.QueryVestingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := k.GetPlan(ctx, req.PlanId)
	if !found {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)

	vested := plan.VestingPlan.VestedAmt(ctx.BlockTime())

	response := &types.QueryVestingResponse{
		Owner:           owner.String(),
		Total:           plan.VestingPlan.Amount,
		VestedAmount:    vested,
		ClaimableAmount: vested.Sub(plan.VestingPlan.Claimed),
	}
	return response, nil
}
