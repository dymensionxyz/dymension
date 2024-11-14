package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) ObsoleteDRSVersions(goCtx context.Context, req *types.QueryObsoleteDRSVersionsRequest) (*types.QueryObsoleteDRSVersionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	versions, err := k.GetAllObsoleteDRSVersions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryObsoleteDRSVersionsResponse{DrsVersions: versions}, nil
}
