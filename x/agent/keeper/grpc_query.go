package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) Agent(goCtx context.Context, req *types.QueryAgentRequest) (*types.QueryAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	agent, err := k.GetAgent(ctx, req.AgentId)
	if errors.Is(err, collections.ErrNotFound) {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, req.AgentId)
	}
	if err != nil {
		return nil, err
	}
	return &types.QueryAgentResponse{Agent: agent}, nil
}
