package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) Agent(goCtx context.Context, req *types.QueryAgentRequest) (*types.QueryAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	agent, found := k.GetAgent(ctx, req.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, req.AgentId)
	}
	fp, err := types.PolicyFingerprint(agent.Policy)
	if err != nil {
		return nil, errorsmod.Wrap(err, "policy fingerprint")
	}
	revoked, err := k.IsPolicyRevoked(ctx, fp)
	if err != nil {
		return nil, errorsmod.Wrap(err, "is policy revoked")
	}
	return &types.QueryAgentResponse{Agent: agent, Fingerprint: fp, Revoked: revoked}, nil
}

func (k Keeper) Agents(goCtx context.Context, req *types.QueryAgentsRequest) (*types.QueryAgentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	agents, pageResp, err := collcompat.CollectionPaginate(ctx, k.agents, req.Pagination,
		func(_ string, a types.Agent) (types.Agent, error) { return a, nil })
	if err != nil {
		return nil, err
	}
	return &types.QueryAgentsResponse{Agents: agents, Pagination: pageResp}, nil
}

func (k Keeper) AgentActions(goCtx context.Context, req *types.QueryAgentActionsRequest) (*types.QueryAgentActionsResponse, error) {
	if req.AgentId == "" {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	actions, pageResp, err := collcompat.CollectionPaginate(ctx, k.actionLog, req.Pagination,
		func(_ collections.Pair[string, uint64], e types.ActionLogEntry) (types.ActionLogEntry, error) {
			return e, nil
		},
		collcompat.WithCollectionPaginationPairPrefix[string, uint64](req.AgentId))
	if err != nil {
		return nil, err
	}
	return &types.QueryAgentActionsResponse{Actions: actions, Pagination: pageResp}, nil
}

func (k Keeper) AgentAction(goCtx context.Context, req *types.QueryAgentActionRequest) (*types.QueryAgentActionResponse, error) {
	if req.AgentId == "" {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	entry, found := k.GetActionLogEntry(ctx, req.AgentId, req.Seq)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrActionNotFound, "agent %s seq %d", req.AgentId, req.Seq)
	}
	return &types.QueryAgentActionResponse{Action: entry}, nil
}

func (k Keeper) RevokedPolicies(goCtx context.Context, _ *types.QueryRevokedPoliciesRequest) (*types.QueryRevokedPoliciesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	fps, err := k.AllRevokedPolicies(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryRevokedPoliciesResponse{Fingerprints: fps}, nil
}

func (k Keeper) PolicyRevoked(goCtx context.Context, req *types.QueryPolicyRevokedRequest) (*types.QueryPolicyRevokedResponse, error) {
	if err := types.ValidateFingerprint(req.Fingerprint); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	revoked, err := k.IsPolicyRevoked(ctx, req.Fingerprint)
	if err != nil {
		return nil, err
	}
	return &types.QueryPolicyRevokedResponse{Revoked: revoked}, nil
}
