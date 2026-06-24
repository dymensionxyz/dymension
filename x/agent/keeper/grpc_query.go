package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	return &types.QueryAgentResponse{Agent: agent}, nil
}
