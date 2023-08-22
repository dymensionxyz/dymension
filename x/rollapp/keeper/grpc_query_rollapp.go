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
		var rollappSummary = types.RollappSummary{
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

		rollapps = append(rollapps, rollappSummary)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRollappResponse{Rollapp: rollapps, Pagination: pageRes}, nil
}

func (k Keeper) Rollapp(c context.Context, req *types.QueryGetRollappRequest) (*types.QueryGetRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetRollapp(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	rollappResponse := &types.QueryGetRollappResponse{Rollapp: val}
	latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, val.RollappId)
	if found {
		rollappResponse.LatestStateIndex = &latestStateInfoIndex
	}
	latestFinalizedStateInfoIndex, found := k.GetLatestFinalizedStateIndex(ctx, val.RollappId)
	if found {
		rollappResponse.LatestFinalizedStateIndex = &latestFinalizedStateInfoIndex
	}

	return rollappResponse, nil
}

func (k Keeper) RollappByEIP155(c context.Context, req *types.QueryGetRollappByEIP155Request) (*types.QueryGetRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetRollappByEIP155(
		ctx,
		req.Eip255,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	rollappResponse := &types.QueryGetRollappResponse{Rollapp: val}
	latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, val.RollappId)
	if found {
		rollappResponse.LatestStateIndex = &latestStateInfoIndex
	}
	latestFinalizedStateInfoIndex, found := k.GetLatestFinalizedStateIndex(ctx, val.RollappId)
	if found {
		rollappResponse.LatestFinalizedStateIndex = &latestFinalizedStateInfoIndex
	}

	return rollappResponse, nil
}
