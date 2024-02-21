package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val := k.GetSequencersByRollapp(
		ctx,
		req.RollappId,
	)
	if len(val) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	var sequencerInfoList []types.Sequencer
	for _, sequencer := range val {
		sequencerInfoList = append(sequencerInfoList, sequencer)
	}

	return &types.QueryGetSequencersByRollappResponse{
		SequencerInfoList: sequencerInfoList,
	}, nil
}

func (k Keeper) SequencersByRollappByStatus(c context.Context, req *types.QueryGetSequencersByRollappByStatusRequest) (*types.QueryGetSequencersByRollappByStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val := k.GetSequencersByRollappByStatus(
		ctx,
		req.RollappId,
		req.Status,
	)
	if len(val) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	var sequencerInfoList []types.Sequencer
	for _, sequencer := range val {
		sequencerInfoList = append(sequencerInfoList, sequencer)
	}

	return &types.QueryGetSequencersByRollappByStatusResponse{
		SequencerInfoList: sequencerInfoList,
	}, nil
}
