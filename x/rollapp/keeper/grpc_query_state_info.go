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

func (k Keeper) StateInfoAll(c context.Context, req *types.QueryAllStateInfoRequest) (*types.QueryAllStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var stateInfos []types.StateInfo
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	stateInfoStore := prefix.NewStore(store, types.KeyPrefix(types.StateInfoKeyPrefix))

	pageRes, err := query.Paginate(stateInfoStore, req.Pagination, func(key []byte, value []byte) error {
		var stateInfo types.StateInfo
		if err := k.cdc.Unmarshal(value, &stateInfo); err != nil {
			return err
		}

		stateInfos = append(stateInfos, stateInfo)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllStateInfoResponse{StateInfo: stateInfos, Pagination: pageRes}, nil
}

func (k Keeper) StateInfo(c context.Context, req *types.QueryGetStateInfoRequest) (*types.QueryGetStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetStateInfo(
		ctx,
		req.RollappId,
		req.StateIndex,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetStateInfoResponse{StateInfo: val}, nil
}
