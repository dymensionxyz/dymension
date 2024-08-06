package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	v2types "github.com/dymensionxyz/dymension/v3/x/rollapp/types/v2"
)

// UpdateState updates the state of a rollapp
func (k msgServerV2) UpdateState(goCtx context.Context, msg *v2types.MsgUpdateState) (*v2types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// load rollapp object for stateful validations
	_, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}

	// call the before-update-state hook
	// checks if:
	// 	1. creator is the sequencer
	// 	2. sequencer rollappId matches the rollappId
	// 	3. sequencer is bonded
	// 	4. sequencer is the proposer
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId)
	if err != nil {
		return nil, err
	}

	// retrieve last updating index
	var newIndex, lastIndex uint64
	latestStateInfoIndex, found := k.GetLatestStateInfoIndex(ctx, msg.RollappId)
	if found {
		// retrieve last updating index
		stateInfo, found := k.GetStateInfo(ctx, msg.RollappId, latestStateInfoIndex.Index)
		// if latestStateInfoIndex exists, there must be an info for this state
		if !found {
			// if not, it's a logic error
			return nil, errorsmod.Wrapf(types.ErrLogic,
				"missing stateInfo for state-index (%d) of rollappId(%s)",
				latestStateInfoIndex.Index, msg.RollappId)
		}

		// check to see if received height is the one we expected
		expectedStartHeight := stateInfo.StartHeight + stateInfo.NumBlocks
		if expectedStartHeight != msg.StartHeight {
			return nil, errorsmod.Wrapf(types.ErrWrongBlockHeight,
				"expected height (%d), but received (%d)",
				expectedStartHeight, msg.StartHeight)
		}

		// bump state index
		lastIndex = latestStateInfoIndex.Index
	}
	newIndex = lastIndex + 1

	daLayer, exists := k.daLayers[msg.DAPath.DaType]
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrInvalidDAClientType, "unknown da layer: %s", msg.DAPath.DaType)
	}
	if err = daLayer.OnRollappStateUpdate(ctx, msg.DAPath.Commitment); err != nil {
		return nil, errorsmod.Wrapf(types.ErrDAClientValidationFailed, "da layer commitment validation failed: %s", err)
	}

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})

	creationHeight := uint64(ctx.BlockHeight())
	daPath, err := msg.DAPath.Marshal()
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidDAClientType, "failed to marshal DAPath: %s", err)
	}
	stateInfo := types.NewStateInfo(msg.RollappId, newIndex, msg.Creator, msg.StartHeight, msg.NumBlocks, string(daPath), creationHeight, msg.BDs)
	// Write new state information to the store indexed by <RollappId,LatestStateInfoIndex>
	k.SetStateInfo(ctx, *stateInfo)

	stateInfoIndex := stateInfo.GetIndex()
	newFinalizationQueue := []types.StateInfoIndex{stateInfoIndex}

	k.Logger(ctx).Debug("Adding state to finalization queue at ", creationHeight)
	// load FinalizationQueue and update
	blockHeightToFinalizationQueue, found := k.GetBlockHeightToFinalizationQueue(ctx, creationHeight)
	if found {
		newFinalizationQueue = append(blockHeightToFinalizationQueue.FinalizationQueue, newFinalizationQueue...)
	}

	// Write new BlockHeightToFinalizationQueue
	k.SetBlockHeightToFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
		CreationHeight:    creationHeight,
		FinalizationQueue: newFinalizationQueue,
	})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStateUpdate,
			stateInfo.GetEvents()...,
		),
	)

	return &v2types.MsgUpdateStateResponse{}, nil
}
