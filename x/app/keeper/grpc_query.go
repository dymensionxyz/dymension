package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) App(goCtx context.Context, req *types.QueryGetAppRequest) (*types.QueryGetAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	app, found := k.GetApp(ctx, req.Name, req.RollappId)
	if !found {
		return nil, types.ErrNotFound
	}

	return &types.QueryGetAppResponse{App: app}, nil
}

func (k Keeper) AppAll(c context.Context, req *types.QueryAllAppRequest) (*types.QueryAllAppResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var apps []types.QueryGetAppResponse
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	rollappStore := prefix.NewStore(store, types.KeyPrefix(types.AppKeyPrefix))

	pageRes, err := query.Paginate(rollappStore, req.Pagination, func(key []byte, value []byte) error {
		var app types.App
		if err := k.cdc.Unmarshal(value, &app); err != nil {
			return err
		}
		apps = append(apps, types.QueryGetAppResponse{App: app})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAppResponse{App: apps, Pagination: pageRes}, nil
}
