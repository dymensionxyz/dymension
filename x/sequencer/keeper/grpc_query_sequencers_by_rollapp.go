package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := k.rollappKeeper.GetRollapp(ctx, req.RollappId); !ok {
		return nil, types.ErrUnknownRollappID
	}

	sequencers := k.GetSequencersByRollapp(ctx, req.RollappId)
	return &types.QueryGetSequencersByRollappResponse{
		Sequencers: sequencers,
	}, nil
}

func (k Keeper) SequencersByRollappByStatus(c context.Context, req *types.QueryGetSequencersByRollappByStatusRequest) (*types.QueryGetSequencersByRollappByStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := k.rollappKeeper.GetRollapp(ctx, req.RollappId); !ok {
		return nil, types.ErrUnknownRollappID
	}

	sequencers := k.GetSequencersByRollappByStatus(
		ctx,
		req.RollappId,
		req.Status,
	)

	return &types.QueryGetSequencersByRollappByStatusResponse{
		Sequencers: sequencers,
	}, nil
}
