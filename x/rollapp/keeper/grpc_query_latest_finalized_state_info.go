package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LatestFinalizedStateInfo(goCtx context.Context, req *types.QueryGetLatestFinalizedStateInfoRequest) (*types.QueryGetLatestFinalizedStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stateInfoIndex, found := k.GetLatestFinalizedStateIndex(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "LatestFinalizedStateIndex not found")
	}

	stateInfo, found := k.GetStateInfo(
		ctx,
		stateInfoIndex.RollappId,
		stateInfoIndex.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "StateInfo not found")
	}

	return &types.QueryGetLatestFinalizedStateInfoResponse{
		StateInfo: stateInfo,
	}, nil
}
