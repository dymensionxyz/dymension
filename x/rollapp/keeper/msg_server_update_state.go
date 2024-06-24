package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateState(goCtx context.Context, msg *types.MsgUpdateState) (*types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	// load rollapp object for stateful validations
	rollapp, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}

	// check rollapp version
	if rollapp.Version != msg.Version {
		return nil, errorsmod.Wrapf(types.ErrVersionMismatch, "rollappId(%s) current version is %d, but got %d", msg.RollappId, rollapp.Version, msg.Version)
	}

	// call the before-update-state hook
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId)
	if err != nil {
		return nil, err
	}

	// Logic Error check - must be done after BeforeUpdateStateRecoverable
	// check if there are permissionedAddresses.
	// if the list is not empty, it means that only premissioned sequencers can be added
	if 0 < len(rollapp.PermissionedAddresses) && !slices.Contains(rollapp.PermissionedAddresses, msg.Creator) {
		// this is a logic error, as the sequencer modules' BeforeUpdateState hook
		// should check that the sequencer exists and register for serving this rollapp
		// so if this check passed, an unpermissioned sequencer is registered
		return nil, errorsmod.Wrapf(types.ErrLogic,
			"unpermissioned sequencer (%s) is registered for rollappId(%s)",
			msg.Creator, msg.RollappId)
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

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})

	creationHeight := uint64(ctx.BlockHeight())
	stateInfo := types.NewStateInfo(msg.RollappId, newIndex, msg.Creator, msg.StartHeight, msg.NumBlocks, msg.DAPath, msg.Version, creationHeight, msg.BDs)
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

	return &types.MsgUpdateStateResponse{}, nil
}
