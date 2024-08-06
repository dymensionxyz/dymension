package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ types.QueryServer = QueryServer{}

type QueryServer struct {
	k Keeper
}

func NewQueryServer(k Keeper) QueryServer {
	return QueryServer{k: k}
}

func (q QueryServer) Params(ctx context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (q QueryServer) Vote(ctx context.Context, request *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (q QueryServer) Distribution(ctx context.Context, request *types.QueryDistributionRequest) (*types.QueryDistributionResponse, error) {
	// TODO implement me
	panic("implement me")
}
