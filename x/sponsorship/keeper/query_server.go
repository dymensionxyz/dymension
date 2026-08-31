package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

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
	params, err := q.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q QueryServer) Vote(goCtx context.Context, request *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	voter, err := sdk.AccAddressFromBech32(request.GetVoter())
	if err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid voter address: %s", err)
	}

	vote, err := q.k.GetVote(ctx, voter)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "vote: %s", err)
		}
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

func (q QueryServer) EstimateClaim(goCtx context.Context, query *types.QueryEstimateClaim) (*types.QueryEstimateClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(query.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid claimer address: %s", err)
	}
	reward, err := q.k.EstimateClaim(ctx, addr, query.RollappId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "claim estimate: %s", err)
		}
		return nil, err
	}
	return &types.QueryEstimateClaimResponse{Reward: reward}, nil
}

func (q QueryServer) Endorsement(goCtx context.Context, query *types.QueryEndorsement) (*types.QueryEndorsementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	endorsement, err := q.k.GetEndorsement(ctx, query.RollappId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "endorsement: %s", err)
		}
		return nil, err
	}
	return &types.QueryEndorsementResponse{Endorsement: endorsement}, nil
}
