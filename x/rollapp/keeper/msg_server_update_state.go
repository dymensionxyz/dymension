package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func (k msgServer) UpdateState(goCtx context.Context, msg *types.MsgUpdateState) (*types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// load rollapp object for stateful validations
	rollapp, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappId
	}

	// check rollapp version
	if rollapp.Version != msg.Version {
		return nil, sdkerrors.Wrapf(types.ErrVersionMismatch, "rollappId(%s) current version is %d, but got %d", msg.RollappId, rollapp.Version, msg.Version)
	}

	// call the before-update-state hook
	err := k.BeforeUpdateStateRecoverable(ctx, msg.Creator, msg.RollappId)
	if err != nil {
		return nil, err
	}

	// Logig Error check - must be done after BeforeUpdateStateRecoverable
	// check if there are permissionedAddresses.
	// if the list is not empty, it means that only premissioned sequencers can be added
	permissionedAddresses := rollapp.PermissionedAddresses.Addresses
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
	stateIndex, isFound := k.GetStateIndex(ctx, msg.RollappId)
	var newIndex uint64
	if !isFound {
		// check to see if it's the first update
		if msg.StartHeight != 0 {
			// if not, it's an error
			return nil, sdkerrors.Wrapf(types.ErrWrongBlockHeight,
				"expected height 0, but received (%d)",
				msg.StartHeight)
		}
		// else, it's the first update
		newIndex = 0
	} else {
		// retrieve last updating index
		stateInfo, isFound := k.GetStateInfo(ctx, msg.RollappId, stateIndex.Index)
		// Check Error: if stateIndex exists, there must me an info for this state
		if !isFound {
			// if not, it's a logic error
			return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"missing stateInfo for state-index (%d) of rollappId(%s)",
				stateIndex.Index, msg.RollappId)
		}

		// check to see if received height is the one we expected
		expectedStartHeight := stateInfo.StartHeight + stateInfo.NumBlocks
		if expectedStartHeight != msg.StartHeight {
			return nil, sdkerrors.Wrapf(types.ErrWrongBlockHeight,
				"expected height (%d), but received (%d)",
				expectedStartHeight, msg.StartHeight)
		}

		// bump state index
		newIndex = stateIndex.Index + 1
	}

	// Write new index information to the store
	k.SetStateIndex(ctx, types.StateIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})

	// Write new state information to the store indexed by <RollappId,StateIndex>
	k.SetStateInfo(ctx, types.StateInfo{
		RollappId:      msg.RollappId,
		StateIndex:     newIndex,
		Sequencer:      msg.Creator,
		StartHeight:    msg.StartHeight,
		NumBlocks:      msg.NumBlocks,
		DAPath:         msg.DAPath,
		Version:        msg.Version,
		CreationHeight: 0, // TODO - have the current block height
		Status:         types.STATE_STATUS_RECEIVED,
		BDs:            msg.BDs},
	)

	return &types.MsgUpdateStateResponse{}, nil
}
