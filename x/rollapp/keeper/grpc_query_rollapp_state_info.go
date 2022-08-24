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

func (k Keeper) RollappStateInfoAll(c context.Context, req *types.QueryAllRollappStateInfoRequest) (*types.QueryAllRollappStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var rollappStateInfos []types.RollappStateInfo
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	rollappStateInfoStore := prefix.NewStore(store, types.KeyPrefix(types.RollappStateInfoKeyPrefix))

	pageRes, err := query.Paginate(rollappStateInfoStore, req.Pagination, func(key []byte, value []byte) error {
		var rollappStateInfo types.RollappStateInfo
		if err := k.cdc.Unmarshal(value, &rollappStateInfo); err != nil {
			return err
		}

		rollappStateInfos = append(rollappStateInfos, rollappStateInfo)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllRollappStateInfoResponse{RollappStateInfo: rollappStateInfos, Pagination: pageRes}, nil
}

func (k Keeper) RollappStateInfo(c context.Context, req *types.QueryGetRollappStateInfoRequest) (*types.QueryGetRollappStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetRollappStateInfo(
		ctx,
		req.RollappId,
		req.StateIndex,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetRollappStateInfoResponse{RollappStateInfo: val}, nil
}
