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

func (k Keeper) SequencerAll(c context.Context, req *types.QueryAllSequencerRequest) (*types.QueryAllSequencerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var sequencerInfoList []types.SequencerInfo
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sequencerStore := prefix.NewStore(store, types.KeyPrefix(types.SequencerKeyPrefix))

	pageRes, err := query.Paginate(sequencerStore, req.Pagination, func(key []byte, value []byte) error {
		var sequencer types.Sequencer
		if err := k.cdc.Unmarshal(value, &sequencer); err != nil {
			return err
		}

		sequencerInfoList = append(sequencerInfoList, types.SequencerInfo{
			Sequencer: sequencer,
			Status:    sequencer.Status,
		})

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSequencerResponse{SequencerInfoList: sequencerInfoList, Pagination: pageRes}, nil
}

func (k Keeper) Sequencer(c context.Context, req *types.QueryGetSequencerRequest) (*types.QueryGetSequencerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	seq, found := k.GetSequencer(
		ctx,
		req.SequencerAddress,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	sequencerInfo := types.SequencerInfo{
		Sequencer: seq,
		Status:    seq.Status,
	}
	return &types.QueryGetSequencerResponse{SequencerInfo: sequencerInfo}, nil
}
