package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	p, err := k.GetParams(sdk.UnwrapSDKContext(c))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryParamsResponse{Params: p}, nil
}

func (k Keeper) Agent(c context.Context, req *types.QueryAgentRequest) (*types.QueryAgentResponse, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	a, err := k.GetAgent(sdk.UnwrapSDKContext(c), req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "agent not found: %s", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAgentResponse{Agent: a}, nil
}

func (k Keeper) Agents(c context.Context, req *types.QueryAgentsRequest) (*types.QueryAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	agents, pageResp, err := k.GetAllAgentsPaginated(sdk.UnwrapSDKContext(c), req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAgentsResponse{Agents: agents, Pagination: pageResp}, nil
}

func (k Keeper) AgentActions(c context.Context, req *types.QueryAgentActionsRequest) (*types.QueryAgentActionsResponse, error) {
	if req == nil || req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	actions, pageResp, err := k.GetAgentActionsPaginated(sdk.UnwrapSDKContext(c), req.AgentId, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAgentActionsResponse{Actions: actions, Pagination: pageResp}, nil
}

func (k Keeper) AgentAction(c context.Context, req *types.QueryAgentActionRequest) (*types.QueryAgentActionResponse, error) {
	if req == nil || req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	e, err := k.GetActionLogEntry(sdk.UnwrapSDKContext(c), req.AgentId, req.Seq)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "action not found: agent %s seq %d", req.AgentId, req.Seq)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAgentActionResponse{Action: e}, nil
}
