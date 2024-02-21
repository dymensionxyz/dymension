package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Sequencers(c context.Context, req *types.QuerySequencersRequest) (*types.QuerySequencersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var sequencers []types.Sequencer
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sequencerStore := prefix.NewStore(store, types.SequencersKey())

	pageRes, err := query.Paginate(sequencerStore, req.Pagination, func(key []byte, value []byte) error {
		var sequencer types.Sequencer
		if err := k.cdc.Unmarshal(value, &sequencer); err != nil {
			return err
		}
		sequencers = append(sequencers, sequencer)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySequencersResponse{Sequencers: sequencers, Pagination: pageRes}, nil
}

func (k Keeper) Sequencer(c context.Context, req *types.QueryGetSequencerRequest) (*types.QueryGetSequencerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	seq, found := k.GetSequencer(ctx, req.SequencerAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSequencerResponse{Sequencer: seq}, nil
}
