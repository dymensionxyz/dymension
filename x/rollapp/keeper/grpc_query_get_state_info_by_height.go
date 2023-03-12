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
	stateInfo, err := k.FindStateInfoByHeight(ctx, req.RollappId, req.Height)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetStateInfoByHeightResponse{StateInfo: *stateInfo}, nil
}
func (k Keeper) FindStateInfoByHeight(ctx sdk.Context, rollappId string, heigh uint64) (*types.StateInfo, error) {
	// check that heigh not zero
	if heigh == 0 {
		return nil, types.ErrInvalidHeight
	}

	_, found := k.GetRollapp(ctx, rollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	stateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"LatestStateInfoIndex wasn't found for rollappId=%s",
			rollappId)
	}
	// initial interval to search in
	startInfoIndex := uint64(1) // see TODO bellow
	endInfoIndex := stateInfoIndex.Index

	// get state info
	LatestStateInfo, found := k.GetStateInfo(ctx, rollappId, endInfoIndex)
	if !found {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
			"StateInfo wasn't found for rollappId=%s, index=%d",
			rollappId, endInfoIndex)
	}

	// check that heigh exists
	if heigh >= LatestStateInfo.StartHeight+LatestStateInfo.NumBlocks {
		return nil, sdkerrors.Wrapf(types.ErrStateNotExists,
			"rollappId=%s, height=%d",
			rollappId, heigh)
	}

	// check if the the heigh belongs to this batch
	if heigh >= LatestStateInfo.StartHeight {
		return &LatestStateInfo, nil
	}

	maxNumberOfSteps := endInfoIndex - startInfoIndex + 1
	stepNum := uint64(0)
	for ; stepNum < maxNumberOfSteps; stepNum += 1 {
		// we know that endInfoIndex > startInfoIndex
		// otherwise the heigh should have been found
		if endInfoIndex <= startInfoIndex {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"endInfoIndex should be != than startInfoIndex rollappId=%s, startInfoIndex=%d, endInfoIndex=%d",
				rollappId, startInfoIndex, endInfoIndex)
		}
		// 1. get state info
		startStateInfo, found := k.GetStateInfo(ctx, rollappId, startInfoIndex)
		if !found {
			// TODO:
			// if stateInfo is missing it won't be logic error if history deletion be implemented
			// for that we will have to check the oldest we have
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for rollappId=%s, index=%d",
				rollappId, startInfoIndex)
		}
		endStateInfo, found := k.GetStateInfo(ctx, rollappId, endInfoIndex)
		if !found {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for rollappId=%s, index=%d",
				rollappId, endInfoIndex)
		}
		startHeight := startStateInfo.StartHeight
		endHeight := endStateInfo.StartHeight + endStateInfo.NumBlocks - 1

		// 2. check startStateInfo
		if heigh >= startStateInfo.StartHeight &&
			(startStateInfo.StartHeight+startStateInfo.NumBlocks) > heigh {
			return &startStateInfo, nil
		}

		// 3. check endStateInfo
		if heigh >= endStateInfo.StartHeight &&
			(endStateInfo.StartHeight+endStateInfo.NumBlocks) > heigh {
			return &endStateInfo, nil
		}

		// 4. calculate the average blocks per batch
		avgBlocksPerBatch := (endHeight - startHeight + 1) / (endInfoIndex - startInfoIndex + 1)
		if avgBlocksPerBatch == 0 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"avgBlocksPerBatch is zero!!! rollappId=%s, endHeight=%d, startHeight=%d, endInfoIndex=%d, startInfoIndex=%d",
				rollappId, endHeight, startHeight, endInfoIndex, startInfoIndex)
		}

		// 5. load the candidate block batch
		infoIndexStep := (heigh - startHeight) / avgBlocksPerBatch
		if infoIndexStep == 0 {
			infoIndexStep = 1
		}
		candidateInfoIndex := startInfoIndex + infoIndexStep
		if candidateInfoIndex > endInfoIndex {
			// skip to the last, probably the steps to big
			candidateInfoIndex = endInfoIndex
		}
		if candidateInfoIndex == endInfoIndex {
			candidateInfoIndex = endInfoIndex - 1
		}
		candidateStateInfo, found := k.GetStateInfo(ctx, rollappId, candidateInfoIndex)
		if !found {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"StateInfo wasn't found for rollappId=%s, index=%d",
				rollappId, candidateInfoIndex)
		}

		// 6. check the candidate
		if candidateStateInfo.StartHeight > heigh {
			endInfoIndex = candidateInfoIndex - 1
		} else {
			if candidateStateInfo.StartHeight+candidateStateInfo.NumBlocks-1 < heigh {
				startInfoIndex = candidateInfoIndex + 1
			} else {
				return &candidateStateInfo, nil
			}
		}
	}

	return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
		"More searching steps than indexes!!! rollappId=%s, stepNum=%d, maxNumberOfSteps=%d",
		rollappId, stepNum, maxNumberOfSteps)
}
