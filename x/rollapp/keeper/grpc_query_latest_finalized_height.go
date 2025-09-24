package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) LatestFinalizedHeight(c context.Context, req *types.QueryGetLatestFinalizedHeightRequest) (*types.QueryGetLatestFinalizedHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	latestFinalizedHeight, err := k.GetLatestFinalizedHeight(ctx, req.RollappId)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetLatestFinalizedHeightResponse{
		Height: latestFinalizedHeight,
	}, nil
}
