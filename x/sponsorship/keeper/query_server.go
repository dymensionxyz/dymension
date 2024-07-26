package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	k Keeper
}

func NewQueryServer(k Keeper) QueryServer {
	return QueryServer{k: k}
}

func (q QueryServer) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q QueryServer) Vote(goCtx context.Context, request *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	voter, err := sdk.AccAddressFromBech32(request.GetVoter())
	if err != nil {
		return nil, fmt.Errorf("invalid voter address: %w", err)
	}

	vote, err := q.k.GetVote(ctx, voter)
	if err != nil {
		return nil, err
	}

	return &types.QueryVoteResponse{Vote: vote}, nil
}

func (q QueryServer) Distribution(goCtx context.Context, _ *types.QueryDistributionRequest) (*types.QueryDistributionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	distribution, err := q.k.GetDistribution(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryDistributionResponse{Distribution: distribution}, nil
}
