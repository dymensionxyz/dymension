package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) StateInfo(c context.Context, req *types.QueryGetStateInfoRequest) (*types.QueryGetStateInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if req.Height == 0 && req.Index == 0 {
		if req.Finalized {
			latestFinalizedStateIndex, found := k.GetLatestFinalizedStateIndex(ctx, req.RollappId)
			if !found {
				return nil, errorsmod.Wrapf(types.ErrNoFinalizedStateYetForRollapp,
					"LatestFinalizedStateIndex wasn't found for rollappId=%s", req.RollappId)
			}
			req.Index = latestFinalizedStateIndex.Index
		} else {
			latestStateIndex, found := k.GetLatestStateInfoIndex(ctx, req.RollappId)
			if !found {
				if _, exists := k.GetRollapp(ctx, req.RollappId); !exists {
					return nil, types.ErrRollappNotRegistered
				}
				return nil, status.Error(codes.NotFound, "not found")
			}
			req.Index = latestStateIndex.Index
		}
	}

	var stateInfo types.StateInfo
	if req.Index != 0 {
		val, found := k.GetStateInfo(ctx, req.RollappId, req.Index)
		if !found {
			return nil, status.Error(codes.NotFound, "not found")
		}
		stateInfo = val
	} else if req.Height != 0 {
		val, err := k.FindStateInfoByHeight(ctx, req.RollappId, req.Height)
		if err != nil {
			return nil, err
		}
		stateInfo = *val
	}

	return &types.QueryGetStateInfoResponse{StateInfo: stateInfo}, nil
}

func (k Keeper) FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*types.StateInfo, error) {
	// check that height not zero
	if height == 0 {
		return nil, types.ErrInvalidHeight
	}

	_, found := k.GetRollapp(ctx, rollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	stateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrNotFound,
			"LatestStateInfoIndex wasn't found for rollappId=%s",
			rollappId)
	}
	// initial interval to search in
	startInfoIndex := uint64(1)
	endInfoIndex := stateInfoIndex.Index
	for startInfoIndex <= endInfoIndex {
		midIndex := startInfoIndex + (endInfoIndex-startInfoIndex)/2
		state, ok := k.GetStateInfo(ctx, rollappId, midIndex)
		if !ok {
			return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "StateInfo wasn't found for rollappId=%s, index=%d", rollappId, midIndex)
		}
		if state.ContainsHeight(height) {
			return &state, nil
		}
		if height < state.GetStartHeight() {
			endInfoIndex = midIndex - 1
		} else {
			startInfoIndex = midIndex + 1
		}
	}
	return nil, errorsmod.Wrapf(types.ErrStateNotExists, "StateInfo wasn't found for rollappId=%s, height=%d", rollappId, height)
}
