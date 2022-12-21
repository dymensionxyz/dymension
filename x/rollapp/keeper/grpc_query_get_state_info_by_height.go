package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetStateInfoByHeight(goCtx context.Context, req *types.QueryGetStateInfoByHeightRequest) (*types.QueryGetStateInfoByHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// check that req.Height not zero
	if req.Height == 0 {
		return nil, types.ErrInvalidHeight
	}

	stateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, req.RollappId)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"LatestStateInfoIndex wasn't found for req.RollappId=%s",
			req.RollappId)
	}
	// initial interval to search in
	startInfoIndex := uint64(1) // see TODO bellow
	endInfoIndex := stateInfoIndex.Index

	// get state info
	LatestStateInfo, found := k.GetStateInfo(ctx, req.RollappId, endInfoIndex)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"StateInfo wasn't found for req.RollappId=%s, index=%d",
			req.RollappId, endInfoIndex)
	}

	// check that req.Height exists
	if req.Height >= LatestStateInfo.StartHeight+LatestStateInfo.NumBlocks {
		return nil, types.ErrStateNotExists
	}

	// check if the the req.Height belongs to this batch
	if req.Height >= LatestStateInfo.StartHeight {
		return &types.QueryGetStateInfoByHeightResponse{StateInfo: LatestStateInfo}, nil
	}

	maxNumberOfSteps := endInfoIndex - startInfoIndex + 1
	stepNum := uint64(0)
	for ; stepNum < maxNumberOfSteps; stepNum += 1 {
		// we know that endInfoIndex > startInfoIndex
		// otherwise the req.Height should have been found
		if endInfoIndex <= startInfoIndex {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"endInfoIndex should be != than startInfoIndex req.RollappId=%s, startInfoIndex=%d, endInfoIndex=%d",
				req.RollappId, startInfoIndex, endInfoIndex)
		}
		// 1. get state info
		startStateInfo, found := k.GetStateInfo(ctx, req.RollappId, startInfoIndex)
		if !found {
			// TODO:
			// if stateInfo is missing it won't be logic error if history deletion be implemented
			// for that we will have to check the oldest we have
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for req.RollappId=%s, index=%d",
				req.RollappId, startInfoIndex)
		}
		endStateInfo, found := k.GetStateInfo(ctx, req.RollappId, endInfoIndex)
		if !found {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for req.RollappId=%s, index=%d",
				req.RollappId, endInfoIndex)
		}
		startHeight := startStateInfo.StartHeight
		endHeight := endStateInfo.StartHeight + endStateInfo.NumBlocks - 1

		// 2. calculate the average blocks per batch
		avgBlocksPerBatch := (endHeight - startHeight + 1) / (endInfoIndex - startInfoIndex + 1)
		if avgBlocksPerBatch == 0 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"avgBlocksPerBatch is zero!!! req.RollappId=%s, endHeight=%d, startHeight=%d, endInfoIndex=%d, startInfoIndex=%d",
				req.RollappId, endHeight, startHeight, endInfoIndex, startInfoIndex)
		}

		// 3. load the candidate block batch
		infoIndexStep := (req.Height - startHeight) / avgBlocksPerBatch
		if infoIndexStep == 0 {
			infoIndexStep = 1
		}
		candidateInfoIndex := startInfoIndex + infoIndexStep
		if candidateInfoIndex > endInfoIndex {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"candidateInfoIndex > endInfoIndex for req.RollappId=%s, endHeight=%d, startHeight=%d, endInfoIndex=%d, startInfoIndex=%d, candidateInfoIndex=%d",
				req.RollappId, endHeight, startHeight, endInfoIndex, startInfoIndex, candidateInfoIndex)
		}
		if candidateInfoIndex == endInfoIndex {
			candidateInfoIndex = endInfoIndex - 1
		}
		candidateStateInfo, found := k.GetStateInfo(ctx, req.RollappId, candidateInfoIndex)
		if !found {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for req.RollappId=%s, index=%d",
				req.RollappId, candidateInfoIndex)
		}

		// 4. check the candidate
		if candidateStateInfo.StartHeight > req.Height {
			startInfoIndex = candidateInfoIndex
		} else {
			if candidateStateInfo.StartHeight+candidateStateInfo.NumBlocks-1 < req.Height {
				endInfoIndex = candidateInfoIndex
			} else {
				return &types.QueryGetStateInfoByHeightResponse{StateInfo: candidateStateInfo}, nil
			}
		}
	}

	return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
		"More searching steps than indexes!!! req.RollappId=%s, stepNum=%d, maxNumberOfSteps=%d",
		req.RollappId, stepNum, maxNumberOfSteps)
}
