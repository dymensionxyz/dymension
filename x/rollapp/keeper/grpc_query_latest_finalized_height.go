package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
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

	val, found := k.GetLatestFinalizedStateIndex(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "latest finalized index")
	}

	state := k.MustGetStateInfo(ctx, req.RollappId, val.Index)

	return &types.QueryGetLatestFinalizedHeightResponse{
		Height: state.GetLatestHeight(),
	}, nil
}