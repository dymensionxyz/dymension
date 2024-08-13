package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
	ctx := sdk.UnwrapSDKContext(c)
	ra, ok := k.GetRollapp(ctx, req.RollappId)
	return getSummaryResponse(ctx, k, ra, ok)
}

func (k Keeper) RollappByEIP155(c context.Context, req *types.QueryGetRollappByEIP155Request) (*types.QueryGetRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	ra, ok := k.GetRollappByEIP155(ctx, req.Eip155)
	return getSummaryResponse(ctx, k, ra, ok)
}

func getSummaryResponse(ctx sdk.Context, k Keeper, rollapp types.Rollapp, ok bool) (*types.QueryGetRollappResponse, error) {
	if !ok {
		return nil, gerrc.ErrNotFound.Wrap("rollapp")
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
