package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LatestStateIndex(c context.Context, req *types.QueryGetLatestStateIndexRequest) (*types.QueryGetLatestStateIndexResponse, error) {
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

	return &types.QueryGetLatestStateIndexResponse{StateIndex: val}, nil
}
