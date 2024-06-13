package keeper

import (
	"context"

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
	statuses := []commontypes.Status{commontypes.Status_PENDING, commontypes.Status_FINALIZED, commontypes.Status_REVERTED}
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
	// Get the demand orders by status, with optional filters
	demandOrders, err := q.ListDemandOrdersByStatus(sdk.UnwrapSDKContext(goCtx), req.Status, int(req.Limit), filterOpts(req)...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Construct the response
	return &types.QueryDemandOrdersByStatusResponse{DemandOrders: demandOrders}, nil
}

func filterOpts(req *types.QueryDemandOrdersByStatusRequest) []filterOption {
	var opts []filterOption
	if req.RollappId != "" {
		opts = append(opts, isRollappId(req.RollappId))
	}
	if req.Type != commontypes.RollappPacket_UNDEFINED {
		opts = append(opts, isOrderType(req.Type))
	}
	if req.FulfillmentState != types.FulfillmentState_UNDEFINED {
		opts = append(opts, isFulfillmentState(req.FulfillmentState))
	}
	return opts
}

type filterOption func(order types.DemandOrder) bool

func isRollappId(rollappId string) filterOption {
	return func(order types.DemandOrder) bool {
		return order.RollappId == rollappId
	}
}

func isOrderType(orderType ...commontypes.RollappPacket_Type) filterOption {
	return func(order types.DemandOrder) bool {
		for _, ot := range orderType {
			if order.Type == ot {
				return true
			}
		}
		return false
	}
}

func isFulfillmentState(fulfillmentState types.FulfillmentState) filterOption {
	return func(order types.DemandOrder) bool {
		return order.IsFulfilled == (types.FulfillmentState_FULFILLED == fulfillmentState)
	}
}
