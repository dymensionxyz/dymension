package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		return nil, sdkerrors.Wrapf(types.ErrVersionMismatch, "rollappId(%s) current version is %d, but got %d", msg.RollappId, rollapp.Version, msg.Version)
	}

	// call the before-update-state hook
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId)
	if err != nil {
		return nil, err
	}

	// Logic Error check - must be done after BeforeUpdateStateRecoverable
	// check if there are permissionedAddresses.
	// if the list is not empty, it means that only premissioned sequencers can be added
	permissionedAddresses := rollapp.PermissionedAddresses
	if len(permissionedAddresses) > 0 {
		bPermissioned := false
		// check to see if the sequencer is in the permissioned list
		for i := range permissionedAddresses {
			if permissionedAddresses[i] == msg.Creator {
				// Found!
				bPermissioned = true
				break
			}
		}
		// Check Error: only permissioned sequencers allowed to update and this one is not in the list
		if !bPermissioned {
			// this is a logic error, as the sequencer modules' BeforeUpdateState hook
			// should check that the sequencer exists and register for serving this rollapp
			// so if this check passed, an unpermissioned sequencer is registered
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"unpermissioned sequencer (%s) is registered for rollappId(%s)",
				msg.Creator, msg.RollappId)
		}
	}

	// retrieve last updating index
	latestStateInfoIndex, isFound := k.GetLatestStateInfoIndex(ctx, msg.RollappId)
	var newIndex uint64
	if !isFound {
		// check to see if it's the first update
		if msg.StartHeight != 1 {
			// if not, it's an error
			return nil, sdkerrors.Wrapf(types.ErrWrongBlockHeight,
				"expected height 1, but received (%d)",
				msg.StartHeight)
		}
		// else, it's the first update
		newIndex = 1
	} else {
		// retrieve last updating index
		stateInfo, isFound := k.GetStateInfo(ctx, msg.RollappId, latestStateInfoIndex.Index)
		// Check Error: if latestStateInfoIndex exists, there must me an info for this state
		if !isFound {
			// if not, it's a logic error
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"missing stateInfo for state-index (%d) of rollappId(%s)",
				latestStateInfoIndex.Index, msg.RollappId)
		}

		// check to see if received height is the one we expected
		expectedStartHeight := stateInfo.StartHeight + stateInfo.NumBlocks
		if expectedStartHeight != msg.StartHeight {
			return nil, sdkerrors.Wrapf(types.ErrWrongBlockHeight,
				"expected height (%d), but received (%d)",
				expectedStartHeight, msg.StartHeight)
		}

		// bump state index
		newIndex = latestStateInfoIndex.Index + 1
	}

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})

	// Write new state information to the store indexed by <RollappId,LatestStateInfoIndex>
	stateInfoIndex := types.StateInfoIndex{RollappId: msg.RollappId, Index: newIndex}
	k.SetStateInfo(ctx, types.StateInfo{
		StateInfoIndex: stateInfoIndex,
		Sequencer:      msg.Creator,
		StartHeight:    msg.StartHeight,
		NumBlocks:      msg.NumBlocks,
		DAPath:         msg.DAPath,
		Version:        msg.Version,
		CreationHeight: uint64(ctx.BlockHeight()),
		Status:         types.STATE_STATUS_RECEIVED,
		BDs:            msg.BDs},
	)

	// calculate finalization
	finalizationHeight := uint64(ctx.BlockHeight()) + k.DisputePeriodInBlocks(ctx)
	newFinalizationQueue := []types.StateInfoIndex{stateInfoIndex}

	// load FinalizationQueue and update
	blockHeightToFinalizationQueue, found := k.GetBlockHeightToFinalizationQueue(ctx, finalizationHeight)
	if found {
		newFinalizationQueue = append(blockHeightToFinalizationQueue.FinalizationQueue, newFinalizationQueue...)
	}

	// Write new BlockHeightToFinalizationQueue
	k.SetBlockHeightToFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
		FinalizationHeight: finalizationHeight,
		FinalizationQueue:  newFinalizationQueue,
	})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStateUpdate,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeyStateInfoIndex, strconv.FormatUint(stateInfoIndex.Index, 10)),
			sdk.NewAttribute(types.AttributeKeyStartHeight, strconv.FormatUint(msg.StartHeight, 10)),
			sdk.NewAttribute(types.AttributeKeyNumBlocks, strconv.FormatUint(msg.NumBlocks, 10)),
			sdk.NewAttribute(types.AttributeKeyDAPath, msg.DAPath),
		),
	)

	return &types.MsgUpdateStateResponse{}, nil
}
