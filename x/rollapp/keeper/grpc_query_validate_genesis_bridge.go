package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) ValidateGenesisBridge(goCtx context.Context, req *types.QueryValidateGenesisBridgeRequest) (*types.QueryValidateGenesisBridgeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	ra, ok := k.GetRollapp(ctx, req.GetRollappId())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, types.ErrRollappNotFound.Error())
	}

	err := types.NewGenesisBridgeValidator(req.Data, ra.GenesisInfo).Validate()
	// we want to distinguish between the gRPC error and the error from the validation,
	// so we put the validation error in the response
	if err != nil {
		return &types.QueryValidateGenesisBridgeResponse{Valid: false, Err: err.Error()}, nil
	}

	return &types.QueryValidateGenesisBridgeResponse{Valid: true}, nil
}
