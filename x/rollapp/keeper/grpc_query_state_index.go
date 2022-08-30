package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LatestStateInfoIndexAll(c context.Context, req *types.QueryAllLatestStateInfoIndexRequest) (*types.QueryAllLatestStateInfoIndexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var latestStateInfoIndexs []types.StateInfoIndex
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	latestStateInfoIndexStore := prefix.NewStore(store, types.KeyPrefix(types.LatestStateInfoIndexKeyPrefix))

	pageRes, err := query.Paginate(latestStateInfoIndexStore, req.Pagination, func(key []byte, value []byte) error {
		var latestStateInfoIndex types.StateInfoIndex
		if err := k.cdc.Unmarshal(value, &latestStateInfoIndex); err != nil {
			return err
		}

		latestStateInfoIndexs = append(latestStateInfoIndexs, latestStateInfoIndex)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLatestStateInfoIndexResponse{LatestStateInfoIndex: latestStateInfoIndexs, Pagination: pageRes}, nil
}

func (k Keeper) LatestStateInfoIndex(c context.Context, req *types.QueryGetLatestStateInfoIndexRequest) (*types.QueryGetLatestStateInfoIndexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetLatestStateInfoIndex(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLatestStateInfoIndexResponse{LatestStateInfoIndex: val}, nil
}
