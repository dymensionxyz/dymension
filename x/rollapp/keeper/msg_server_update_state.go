package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateState(goCtx context.Context, msg *types.MsgUpdateState) (*types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// load rollapp object for stateful validations
	rollapp, isFound := k.GetRollapp(ctx, msg.RollappId)
	if !isFound {
		return nil, types.ErrUnknownRollappID
	}

	// verify the rollapp is not frozen
	if rollapp.Frozen {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp is frozen")
	}

	// call the before-update-state hook
	// currently used by `x/sequencer` to:
	// 1. validate the state update submitter
	// 2. complete the rotation of the proposer if needed
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId, msg.Last)
	if err != nil {
		return nil, errorsmod.Wrap(err, "before update state")
	}

	// verify the DRS version is not vulnerable
	if k.IsDRSVersionVulnerable(ctx, msg.DrsVersion) {
		// the rollapp is not marked as vulnerable yet, mark it now
		err := k.MarkRollappAsVulnerable(ctx, msg.RollappId)
		if err != nil {
			return nil, fmt.Errorf("mark rollapp vulnerable: %w", err)
		}
		k.Logger(ctx).With("rollapp_id", msg.RollappId, "drs_version", msg.DrsVersion).
			Info("non-frozen rollapp tried to submit MsgUpdateState with the vulnerable DRS version, mark the rollapp as vulnerable")
		// we must return non-error if we want the changes to be saved
		return &types.MsgUpdateStateResponse{}, nil
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

		// if previous block descriptor has timestamp, it means the rollapp is upgraded
		// therefore all new BDs need to have timestamp
		lastBD := stateInfo.GetLatestBlockDescriptor()
		if !lastBD.Timestamp.IsZero() {
			err := msg.BDs.Validate()
			if err != nil {
				return nil, errorsmod.Wrap(err, "block descriptors")
			}
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
	} else {
		err := msg.BDs.Validate()
		if err != nil {
			return nil, errorsmod.Wrap(err, "block descriptors")
		}
	}
	newIndex = lastIndex + 1

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})

	creationHeight := uint64(ctx.BlockHeight())
	blockTime := ctx.BlockTime()
	stateInfo := types.NewStateInfo(
		msg.RollappId,
		newIndex,
		msg.Creator,
		msg.StartHeight,
		msg.NumBlocks,
		msg.DAPath,
		creationHeight,
		msg.BDs,
		blockTime,
		msg.DrsVersion,
	)
	// Write new state information to the store indexed by <RollappId,LatestStateInfoIndex>
	k.SetStateInfo(ctx, *stateInfo)

	err = k.hooks.AfterUpdateState(ctx, msg.RollappId, stateInfo)
	if err != nil {
		return nil, errorsmod.Wrap(err, "after update state")
	}

	stateInfoIndex := stateInfo.GetIndex()
	newFinalizationQueue := []types.StateInfoIndex{stateInfoIndex}

	k.Logger(ctx).Debug("Adding state to finalization queue at %d", creationHeight)
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

	// TODO: enforce `final_state_update_timeout` if sequencer rotation is in progress
	// https://github.com/dymensionxyz/dymension/issues/1085
	k.IndicateLiveness(ctx, &rollapp)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStateUpdate,
			stateInfo.GetEvents()...,
		),
	)

	return &types.MsgUpdateStateResponse{}, nil
}
