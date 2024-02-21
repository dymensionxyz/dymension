package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	sequencers := k.GetSequencersByRollapp(ctx, req.RollappId)
	if len(sequencers) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSequencersByRollappResponse{
		Sequencers: sequencers,
	}, nil
}

func (k Keeper) SequencersByRollappByStatus(c context.Context, req *types.QueryGetSequencersByRollappByStatusRequest) (*types.QueryGetSequencersByRollappByStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	sequencers := k.GetSequencersByRollappByStatus(
		ctx,
		req.RollappId,
		req.Status,
	)
	if len(sequencers) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSequencersByRollappByStatusResponse{
		Sequencers: sequencers,
	}, nil
}
