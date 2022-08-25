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

func (k Keeper) StateIndexAll(c context.Context, req *types.QueryAllStateIndexRequest) (*types.QueryAllStateIndexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var stateIndexs []types.StateIndex
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	stateIndexStore := prefix.NewStore(store, types.KeyPrefix(types.StateIndexKeyPrefix))

	pageRes, err := query.Paginate(stateIndexStore, req.Pagination, func(key []byte, value []byte) error {
		var stateIndex types.StateIndex
		if err := k.cdc.Unmarshal(value, &stateIndex); err != nil {
			return err
		}

		stateIndexs = append(stateIndexs, stateIndex)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllStateIndexResponse{StateIndex: stateIndexs, Pagination: pageRes}, nil
}

func (k Keeper) StateIndex(c context.Context, req *types.QueryGetStateIndexRequest) (*types.QueryGetStateIndexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetStateIndex(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetStateIndexResponse{StateIndex: val}, nil
}
