package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LatestHeight(c context.Context, req *types.QueryGetLatestHeightRequest) (*types.QueryGetLatestHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var val types.StateInfoIndex
	var found bool
	if req.Finalized {
		val, found = k.GetLatestFinalizedStateIndex(
			ctx,
			req.RollappId,
		)
	} else {
		val, found = k.GetLatestStateInfoIndex(
			ctx,
			req.RollappId,
		)
	}
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	state := k.MustGetStateInfo(ctx, req.RollappId, val.Index)

	return &types.QueryGetLatestHeightResponse{
		Height: state.GetLatestHeight(),
	}, nil
}
