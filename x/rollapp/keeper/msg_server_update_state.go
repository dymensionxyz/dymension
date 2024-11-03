package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	// currently used by `x/sequencer` to:
	// 1. validate the state update submitter
	// 2. complete the rotation of the proposer if needed
	err := k.hooks.BeforeUpdateState(ctx, msg.Creator, msg.RollappId, msg.Last)
	if err != nil {
		return nil, errorsmod.Wrap(err, "before update state")
	}

	// validate correct rollapp revision number
	if rollapp.RevisionNumber != msg.RollappRevision {
		return nil, errorsmod.Wrapf(types.ErrWrongRollappRevision,
			"expected revision number (%d), but received (%d)",
			rollapp.RevisionNumber, msg.RollappRevision)
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

	// it takes the actual proposer because the next one have already been set
	// by the sequencer rotation in k.hooks.BeforeUpdateState
	// the proposer we get is the one that will propose the next block.
	val, _ := k.sequencerKeeper.GetProposer(ctx, msg.RollappId)

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
		val.Address,
	)

	// verify the DRS version is not vulnerable
	// check only last block descriptor DRS, since if that last is not vulnerable it means the rollapp already upgraded and is not vulnerable anymore
	// Rollapp is using a vulnerable DRS version, hard fork it
	if k.IsStateUpdateVulnerable(ctx, stateInfo) {
		err := k.HardForkToLatest(ctx, msg.RollappId)
		if err != nil {
			return nil, fmt.Errorf("mark rollapp vulnerable: %w", err)
		}
		k.Logger(ctx).With("rollapp_id", msg.RollappId, "drs_version", stateInfo.GetLatestBlockDescriptor().DrsVersion).
			Info("rollapp tried to submit MsgUpdateState with the vulnerable DRS version, mark the rollapp as vulnerable")

		// we must return non-error if we want the changes to be saved
		return &types.MsgUpdateStateResponse{}, nil
	}

	// Write new index information to the store
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: msg.RollappId,
		Index:     newIndex,
	})
	// Write new state information to the store indexed by <RollappId,LatestStateInfoIndex>
	k.SetStateInfo(ctx, *stateInfo)

	// call the after-update-state hook
	// currently used by `x/lightclient` to validate the state update in regards to the light client
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

	events := stateInfo.GetEvents()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeStateUpdate,
			events...,
		),
	)

	return &types.MsgUpdateStateResponse{}, nil
}

// IsStateUpdateVulnerable checks if the given DRS version is vulnerable
func (k msgServer) IsStateUpdateVulnerable(ctx sdk.Context, stateInfo *types.StateInfo) bool {
	return k.IsDRSVersionVulnerable(ctx, stateInfo.GetLatestBlockDescriptor().DrsVersion)
}
