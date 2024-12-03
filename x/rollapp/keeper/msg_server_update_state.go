package keeper

import (
	"context"

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

	// call the before-update-state hook
	// currently used by `x/sequencer` to validate the proposer
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId, msg.Last)
	if err != nil {
		return nil, errorsmod.Wrap(err, "before update state")
	}

	// validate correct rollapp revision number
	if rollapp.LatestRevision().Number != msg.RollappRevision {
		return nil, errorsmod.Wrapf(types.ErrWrongRollappRevision,
			"expected revision number (%d), but received (%d)",
			rollapp.LatestRevision().Number, msg.RollappRevision)
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

	// if no rotation, next is the same!
	successor := k.SequencerK.GetProposer(ctx, msg.RollappId)
	if msg.Last {
		successor = k.SequencerK.GetSuccessor(ctx, msg.RollappId)
	}

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
		successor.Address,
	)

	// verify the DRS version is not obsolete
	// check only last block descriptor DRS, since if that last is not obsolete it means the rollapp already upgraded and is not obsolete anymore
	if k.IsStateUpdateObsolete(ctx, stateInfo) {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "MsgUpdateState with an obsolete DRS version. rollapp_id: %s, drs_version: %d",
			msg.RollappId, stateInfo.GetLatestBlockDescriptor().DrsVersion)
	}

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})
	// Write new state information to the store indexed by <RollappId,LatestStateInfoIndex>
	k.SetStateInfo(ctx, *stateInfo)

	// call the after-update-state hook
	// currently used by `x/lightclient` to validate the state update against consensus states
	// x/sequencer will complete the rotation if needed
	err = k.hooks.AfterUpdateState(ctx, &types.StateInfoMeta{
		StateInfo: *stateInfo,
		Revision:  msg.RollappRevision,
		Rollapp:   msg.RollappId,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "hook: after update state")
	}

	stateInfoIndex := stateInfo.GetIndex()
	newFinalizationQueue := []types.StateInfoIndex{stateInfoIndex}

	k.Logger(ctx).Debug("Adding state to finalization queue at %d", creationHeight)

	// load FinalizationQueue and update
	finalizationQueue, found := k.GetFinalizationQueue(ctx, creationHeight, msg.RollappId)
	if found {
		newFinalizationQueue = append(finalizationQueue.FinalizationQueue, newFinalizationQueue...)
	}

	// Write new BlockHeightToFinalizationQueue
	err = k.SetFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
		CreationHeight:    creationHeight,
		FinalizationQueue: newFinalizationQueue,
		RollappId:         msg.RollappId,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "set finalization queue")
	}

	// FIXME: only single save can be done with the latest height
	for _, bd := range msg.BDs.BD {
		if err := k.SaveSequencerHeight(ctx, stateInfo.Sequencer, bd.Height); err != nil {
			return nil, errorsmod.Wrap(err, "save sequencer height")
		}
	}

	// TODO: enforce `final_state_update_timeout` if sequencer rotation is in progress
	// https://github.com/dymensionxyz/dymension/issues/1085
	rollapp = k.MustGetRollapp(ctx, msg.RollappId)
	k.IndicateLiveness(ctx, &rollapp)
	k.SetRollapp(ctx, rollapp)

	events := stateInfo.GetEvents()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStateUpdate,
			events...,
		),
	)

	return &types.MsgUpdateStateResponse{}, nil
}

// IsStateUpdateObsolete checks if the given DRS version is obsolete
func (k msgServer) IsStateUpdateObsolete(ctx sdk.Context, stateInfo *types.StateInfo) bool {
	return k.IsDRSVersionObsolete(ctx, stateInfo.GetLatestBlockDescriptor().DrsVersion)
}
