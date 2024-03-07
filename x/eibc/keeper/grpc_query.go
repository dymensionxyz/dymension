package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

func (q Querier) DemandOrderById(goCtx context.Context, req *types.QueryGetDemandOrderRequest) (*types.QueryGetDemandOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the demand order by its ID and search for it in all statuses
	var demandOrder *types.DemandOrder
	var err error
	var statuses = []commontypes.Status{commontypes.Status_PENDING, commontypes.Status_FINALIZED, commontypes.Status_REVERTED}
	for _, status := range statuses {
		demandOrder, err = q.GetDemandOrder(ctx, status, req.Id)
		if err == nil && demandOrder != nil {
			return &types.QueryGetDemandOrderResponse{DemandOrder: demandOrder}, nil
		}
	}
	return nil, status.Error(codes.Internal, err.Error())
}

func (q Querier) DemandOrdersByStatus(goCtx context.Context, req *types.QueryDemandOrdersByStatusRequest) (*types.QueryDemandOrdersByStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "status must be provided")
	}

	// Convert string status to commontypes.Status
	var statusValue commontypes.Status
	switch strings.ToUpper(req.Status) {
	case "PENDING":
		statusValue = commontypes.Status_PENDING
	case "FINALIZED":
		statusValue = commontypes.Status_FINALIZED
	case "REVERTED":
		statusValue = commontypes.Status_REVERTED
	default:
		return nil, fmt.Errorf("invalid demand order status: %s", req.Status)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the demand orders by status
	demandOrders, err := q.ListDemandOrdersByStatus(ctx, statusValue)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Construct the response
	return &types.QueryDemandOrdersByStatusResponse{DemandOrders: demandOrders}, nil
}
