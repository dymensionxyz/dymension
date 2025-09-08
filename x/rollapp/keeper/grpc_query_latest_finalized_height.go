package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) LatestFinalizedHeight(c context.Context, req *types.QueryGetLatestFinalizedHeightRequest) (*types.QueryGetLatestFinalizedHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	latestIndex, found := k.GetLatestFinalizedStateIndex(ctx, req.RollappId)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("latest finalized state index is not found")
	}

	stateInfo := k.MustGetStateInfo(ctx, req.RollappId, latestIndex.Index)

	return &types.QueryGetLatestFinalizedHeightResponse{
		Height: stateInfo.GetLatestHeight(),
	}, nil
}
