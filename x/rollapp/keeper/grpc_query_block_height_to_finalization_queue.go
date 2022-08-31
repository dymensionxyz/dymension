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

func (k Keeper) BlockHeightToFinalizationQueueAll(c context.Context, req *types.QueryAllBlockHeightToFinalizationQueueRequest) (*types.QueryAllBlockHeightToFinalizationQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blockHeightToFinalizationQueues []types.BlockHeightToFinalizationQueue
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	blockHeightToFinalizationQueueStore := prefix.NewStore(store, types.KeyPrefix(types.BlockHeightToFinalizationQueueKeyPrefix))

	pageRes, err := query.Paginate(blockHeightToFinalizationQueueStore, req.Pagination, func(key []byte, value []byte) error {
		var blockHeightToFinalizationQueue types.BlockHeightToFinalizationQueue
		if err := k.cdc.Unmarshal(value, &blockHeightToFinalizationQueue); err != nil {
			return err
		}

		blockHeightToFinalizationQueues = append(blockHeightToFinalizationQueues, blockHeightToFinalizationQueue)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBlockHeightToFinalizationQueueResponse{BlockHeightToFinalizationQueue: blockHeightToFinalizationQueues, Pagination: pageRes}, nil
}

func (k Keeper) BlockHeightToFinalizationQueue(c context.Context, req *types.QueryGetBlockHeightToFinalizationQueueRequest) (*types.QueryGetBlockHeightToFinalizationQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetBlockHeightToFinalizationQueue(
		ctx,
		req.FinalizationHeight,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBlockHeightToFinalizationQueueResponse{BlockHeightToFinalizationQueue: val}, nil
}
