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

func (k Keeper) SequencersByRollappAll(c context.Context, req *types.QueryAllSequencersByRollappRequest) (*types.QueryAllSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var sequencersByRollapps []types.SequencersByRollapp
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sequencersByRollappStore := prefix.NewStore(store, types.KeyPrefix(types.SequencersByRollappKeyPrefix))

	pageRes, err := query.Paginate(sequencersByRollappStore, req.Pagination, func(key []byte, value []byte) error {
		var sequencersByRollapp types.SequencersByRollapp
		if err := k.cdc.Unmarshal(value, &sequencersByRollapp); err != nil {
			return err
		}

		sequencersByRollapps = append(sequencersByRollapps, sequencersByRollapp)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSequencersByRollappResponse{SequencersByRollapp: sequencersByRollapps, Pagination: pageRes}, nil
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

	return &types.QueryGetSequencersByRollappResponse{SequencersByRollapp: val}, nil
}
