package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SequencersByRollappAll(c context.Context, req *types.QueryAllSequencersByRollappRequest) (*types.QueryAllSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var sequencersByRollappList []types.QueryGetSequencersByRollappResponse
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sequencersByRollappStore := prefix.NewStore(store, types.KeyPrefix(types.SequencersByRollappKeyPrefix))
	sequencerStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))

	pageRes, err := query.Paginate(sequencersByRollappStore, req.Pagination, func(key []byte, value []byte) error {
		var sequencersByRollapp types.SequencersByRollapp
		if err := k.cdc.Unmarshal(value, &sequencersByRollapp); err != nil {
			return err
		}

		var sequencerInfoList []types.SequencerInfo
		for _, sequencerAddress := range sequencersByRollapp.Sequencers.Addresses {
			sequencerVal := sequencerStore.Get(types.SequencerKey(
				sequencerAddress,
			))
			if sequencerVal == nil {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic,
					"sequencer was not found for address %s", sequencerAddress)

			}

			var sequencer types.Sequencer
			k.cdc.MustUnmarshal(sequencerVal, &sequencer)

			sequencerInfoList = append(sequencerInfoList, types.SequencerInfo{
				Sequencer: sequencer,
				Status:    sequencer.Status,
			})
		}

		sequencersByRollappList = append(sequencersByRollappList, types.QueryGetSequencersByRollappResponse{
			RollappId:         sequencersByRollapp.RollappId,
			SequencerInfoList: sequencerInfoList,
		})

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSequencersByRollappResponse{SequencersByRollapp: sequencersByRollappList, Pagination: pageRes}, nil
}

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetSequencersByRollapp(
		ctx,
		req.RollappId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	sequencerStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))

	var sequencerInfoList []types.SequencerInfo
	for _, sequencerAddress := range val.Sequencers.Addresses {

		sequencerVal := sequencerStore.Get(types.SequencerKey(
			sequencerAddress,
		))
		if sequencerVal == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"sequencer was not found for address %s", sequencerAddress)

		}

		var sequencer types.Sequencer
		k.cdc.MustUnmarshal(sequencerVal, &sequencer)

		sequencerInfoList = append(sequencerInfoList, types.SequencerInfo{
			Sequencer: sequencer,
			Status:    sequencer.Status,
		})
	}

	return &types.QueryGetSequencersByRollappResponse{
		RollappId:         req.RollappId,
		SequencerInfoList: sequencerInfoList,
	}, nil
}
