package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SchedulerAll(c context.Context, req *types.QueryAllSchedulerRequest) (*types.QueryAllSchedulerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var schedulers []types.Scheduler
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	schedulerStore := prefix.NewStore(store, types.KeyPrefix(types.SchedulerKeyPrefix))

	pageRes, err := query.Paginate(schedulerStore, req.Pagination, func(key []byte, value []byte) error {
		var scheduler types.Scheduler
		if err := k.cdc.Unmarshal(value, &scheduler); err != nil {
			return err
		}

		schedulers = append(schedulers, scheduler)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSchedulerResponse{Scheduler: schedulers, Pagination: pageRes}, nil
}

func (k Keeper) Scheduler(c context.Context, req *types.QueryGetSchedulerRequest) (*types.QueryGetSchedulerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetScheduler(
		ctx,
		req.SequencerAddress,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSchedulerResponse{Scheduler: val}, nil
}
