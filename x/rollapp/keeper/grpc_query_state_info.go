package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
