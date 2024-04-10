package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

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

// GetPackets implements types.QueryServer.
func (q Querier) GetPackets(goCtx context.Context, req *types.QueryRollappPacketsRequest) (*types.QueryRollappPacketListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	res := &types.QueryRollappPacketListResponse{}

	if req.RollappId == "" {
		// query by status
		res.RollappPackets = q.ListRollappPackets(ctx, types.ByStatus(req.Status))
	} else {
		// query by rollapp id
		res.RollappPackets = q.ListRollappPackets(ctx, types.ByRollappIDByStatus(req.RollappId, req.Status))
	}

	// TODO: handle pagination

	return res, nil
}
