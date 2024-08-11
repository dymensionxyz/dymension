package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) RollappAll(c context.Context, req *types.QueryAllRollappRequest) (*types.QueryAllRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var rollapps []types.RollappSummary
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	rollappStore := prefix.NewStore(store, types.KeyPrefix(types.RollappKeyPrefix))

	pageRes, err := query.Paginate(rollappStore, req.Pagination, func(key []byte, value []byte) error {
		var rollapp types.Rollapp
		if err := k.cdc.Unmarshal(value, &rollapp); err != nil {
			return err
		}
		rollapps = append(rollapps, k.buildRollappSummary(ctx, rollapp))
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRollappResponse{Rollapp: rollapps, Pagination: pageRes}, nil
}

func (k Keeper) Rollapp(c context.Context, req *types.QueryGetRollappRequest) (*types.QueryGetRollappResponse, error) {
	return queryRollapp[types.QueryGetRollappRequest](c, k, req, k.GetRollapp, req.GetRollappId)
}

func (k Keeper) RollappByEIP155(c context.Context, req *types.QueryGetRollappByEIP155Request) (*types.QueryGetRollappResponse, error) {
	return queryRollapp[types.QueryGetRollappByEIP155Request](c, k, req, k.GetRollappByEIP155, req.GetEip155)
}

type (
	queryRollappFn[T any]       func(ctx sdk.Context, q T) (val types.Rollapp, found bool)
	queryRollappGetArgFn[T any] func() T
)

func queryRollapp[Q, T any](c context.Context, k Keeper, req any, qFn queryRollappFn[T], argFn queryRollappGetArgFn[T]) (*types.QueryGetRollappResponse, error) {
	if req == (*Q)(nil) {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	rollapp, found := qFn(ctx, argFn())
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	summary := k.buildRollappSummary(ctx, rollapp)

	return &types.QueryGetRollappResponse{
		Rollapp:                   rollapp,
		LatestStateIndex:          summary.LatestStateIndex,
		LatestFinalizedStateIndex: summary.LatestFinalizedStateIndex,
	}, nil
}

func (k Keeper) buildRollappSummary(ctx sdk.Context, rollapp types.Rollapp) types.RollappSummary {
	rollappSummary := types.RollappSummary{
		RollappId: rollapp.RollappId,
	}
	latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, rollapp.RollappId)
	if found {
		rollappSummary.LatestStateIndex = &latestStateInfoIndex
	}
	latestFinalizedStateInfoIndex, found := k.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
	if found {
		rollappSummary.LatestFinalizedStateIndex = &latestFinalizedStateInfoIndex
	}
	return rollappSummary
}
