package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	cometbfttypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// HardFork handles the fraud evidence submitted by the user.
func (k Keeper) HardFork(ctx sdk.Context, rollappID, clientID string, fraudHeight uint64, seqAddr string) error {
	// validate the request
	err := k.validateForkRequest(ctx, rollappID, clientID, fraudHeight, seqAddr)
	if err != nil {
		return err
	}

	// check height is not finalized
	if stateInfo.Status == common.Status_FINALIZED {
		return errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", fraudHeight)
	}

	// check height is not reverted
	if stateInfo.Status == common.Status_REVERTED {
		return errorsmod.Wrapf(types.ErrDisputeAlreadyReverted, "state info for height %d is already reverted", fraudHeight)
	}

	// check the sequencer for this height is the same as the one in the fraud evidence
	if stateInfo.Sequencer != seqAddr {
		return errorsmod.Wrapf(types.ErrWrongProposerAddr, "sequencer address %s does not match the one in the state info", seqAddr)
	}

	// check that the clientID is correct
	err = k.verifyClientID(ctx, rollappID, clientID)
	if err != nil {
		return errors.Join(types.ErrWrongClientId, err)
	}

	// freeze the rollapp and revert all pending states
	err = k.FreezeRollapp(ctx, rollappID)
	if err != nil {
		return fmt.Errorf("freeze rollapp: %w", err)
	}

	// slash the sequencer, clean delayed packets
	err = k.hooks.OnHardFork(ctx, rollappID, fraudHeight, seqAddr)
	if err != nil {
		return err
	}

	// get current sequencer; might be sentinel.
	currentSequencer := k.sequencersKeeper.GetCurrentRollapSequencer(ctx, rollappID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraud,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
			sdk.NewAttribute(types.AttributeKeyFraudHeight, fmt.Sprint(fraudHeight)),
			sdk.NewAttribute(types.AttributeKeyFraudSequencer, seqAddr),
			sdk.NewAttribute(types.AttributeKeyClientID, clientID),
		),
	)

	return nil
}

// validateForkRequest validates the request for a hard fork.
func (k Keeper) validateForkRequest(ctx sdk.Context, rollappID, clientID string, fraudHeight uint64, seqAddr string) error {
	// check the rollapp exists
	_, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return errorsmod.Wrapf(types.ErrUnknownRollappID, "rollapp with ID %s not found", rollappID)
	}

	// check the clientID is correct
	err := k.verifyClientID(ctx, rollappID, clientID)
	if err != nil {
		return errors.Join(types.ErrWrongClientId, err)
	}

	// check the sequencer is the same as the one in the state info
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, fraudHeight)
	if err != nil {
		return err
	}
	if stateInfo.Sequencer != seqAddr {
		return errorsmod.Wrapf(types.ErrWrongProposerAddr, "sequencer address %s does not match the one in the state info", seqAddr)
	}

	/*
		stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, fraudHeight)
		if err != nil {
			return err
		}

		// check height is not finalized
		if stateInfo.Status == common.Status_FINALIZED {
			return errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", fraudHeight)
		}

		// check height is not reverted
		if stateInfo.Status == common.Status_REVERTED {
			return errorsmod.Wrapf(types.ErrDisputeAlreadyReverted, "state info for height %d is already reverted", fraudHeight)
		}

		// check the sequencer for this height is the same as the one in the fraud evidence
		if stateInfo.Sequencer != seqAddr {
			return errorsmod.Wrapf(types.ErrWrongProposerAddr, "sequencer address %s does not match the one in the state info", seqAddr)
		}

		// check that the clientID is correct
		err = k.verifyClientID(ctx, rollappID, clientID)
		if err != nil {
			return errors.Join(types.ErrWrongClientId, err)
		}
	*/

	return nil
}

// FreezeRollapp marks the rollapp as frozen and reverts all pending states.
// NB! This method is going to be changed as soon as the "Freezing" ADR is ready.
func (k Keeper) FreezeRollapp(ctx sdk.Context, rollappID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	/*
			// remove state updates
		k.removeStateUpdatesUntil(ctx, rollappID, height+1)

		// update revision number
		k.modifyRollappForRollback(ctx, &rollapp, mustUpgrade)
	*/

	rollapp.Frozen = true

	k.RevertPendingStates(ctx, rollappID)

	if rollapp.ChannelId != "" {
		clientID, _, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", rollapp.ChannelId)
		if err != nil {
			return fmt.Errorf("get channel client state: %w", err)
		}

		err = k.freezeClientState(ctx, clientID)
		if err != nil {
			return fmt.Errorf("freeze client state: %w", err)
		}
	}

	k.SetRollapp(ctx, rollapp)
	return nil
}

// freeze IBC client state
func (k Keeper) freezeClientState(ctx sdk.Context, clientId string) error {
	clientState, ok := k.ibcClientKeeper.GetClientState(ctx, clientId)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidClientState, "client state for clientID %s not found", clientId)
	}

	tmClientState, ok := clientState.(*cometbfttypes.ClientState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidClientState, "client state with ID %s is not a tendermint client state", clientId)
	}

	tmClientState.FrozenHeight = clienttypes.NewHeight(tmClientState.GetLatestHeight().GetRevisionHeight(), tmClientState.GetLatestHeight().GetRevisionNumber())
	k.ibcClientKeeper.SetClientState(ctx, clientId, tmClientState)

	return nil
}

// revert all pending states of a rollapp
func (k Keeper) RevertPendingStates(ctx sdk.Context, rollappID string) {
	// TODO (#631): Prefix store by rollappID for efficient querying
	queuePerHeight := k.GetAllBlockHeightToFinalizationQueue(ctx)
	for _, queue := range queuePerHeight {
		leftPendingStates := []types.StateInfoIndex{}
		for _, stateInfoIndex := range queue.FinalizationQueue {
			// keep pending packets not related to this rollapp in the queue
			if stateInfoIndex.RollappId != rollappID {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}

			stateInfo, _ := k.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			stateInfo.Status = common.Status_REVERTED
			k.SetStateInfo(ctx, stateInfo)
		}

		if len(leftPendingStates) == 0 {
			k.RemoveBlockHeightToFinalizationQueue(ctx, queue.CreationHeight)
		} else {
			k.SetBlockHeightToFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
				CreationHeight:    queue.CreationHeight,
				FinalizationQueue: leftPendingStates,
			})
		}
	}
}

// verifyClientID verifies that the provided clientID is the same clientID used by the provided rollapp.
// Possible scenarios:
//  1. both channelID and clientID are empty -> okay
//  2. channelID is empty while clientID is not -> error: rollapp does not have a channel
//  3. clientID is empty while channelID is not -> error: rollapp does have a channel, but the provided clientID is empty
//  4. both channelID and clientID are not empty -> okay: compare the provided channelID against the one from IBC
func (k Keeper) verifyClientID(ctx sdk.Context, rollappID, clientID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	var (
		emptyRollappChannelID = rollapp.ChannelId == ""
		emptyClientID         = clientID == ""
	)

	switch {
	// both channelID and clientID are empty
	case emptyRollappChannelID && emptyClientID:
		return nil // everything is fine, expected scenario

	// channelID is empty while clientID is not
	case emptyRollappChannelID:
		return fmt.Errorf("rollapp does not have a channel: rollapp '%s'", rollappID)

	// clientID is empty while channelID is not
	case emptyClientID:
		return fmt.Errorf("empty clientID while the rollapp channelID is not empty")

	// both channelID and clientID are not empty
	default:
		// extract rollapp channelID
		extractedClientId, _, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", rollapp.ChannelId)
		if err != nil {
			return fmt.Errorf("get channel client state: %w", err)
		}
		// compare it with the passed clientID
		if extractedClientId != clientID {
			return fmt.Errorf("clientID does not match the one in the rollapp: clientID %s, rollapp clientID %s", clientID, extractedClientId)
		}
		return nil
	}
}

/*

// removes state updates until the one specified and included
func (k Keeper) removeStateUpdatesUntil(ctx, rollappID, height) {
	batch, found := k.getBatchID(ctx, rollappID, height)
	// if batch not found we revert nothing, no rollbacks fundamentally.
	if !found {
		return
	}

	lastBatch := k.lastBatch(ctx, rollappID)
	// remove batches until the rollback one
	for cursor := batch.Id; cursor <= lastBatch.Id; cursor++ {
		k.removeBatch(ctx, rollappID, cursor)
	}
	// remove block descriptors until the one we need to rollback to.
	for {
	   bd := batch.BDs.Pop()
	   if bd.Height < height {
		   break
	   }
	}
	if batch.BDs.len() == 0 {
	    // no more blocks in the faulty batch means remove it
	    k.removeBatch(ctx, rollappID, batch.Id)
	} else {
	   // this means we need to truncate the batch
	   // which also becomes the last batch.
	   k.setBatch(ctx, rollappID, batch)
	}
}



func (k Keeper) modifyRollappForRollback(ctx, *rollapp, mustUpgrade) {
   rollapp.RevisionNumber += 1
   rollapp.LastHubUpdate = ctx.Height() // marks the last update done by the hub
   // other things are needed for liveness slashing, refer to the spec: https://www.notion.so/dymension/sequencer-jailing-slashing-3455fe70923143cbbfd8f96d71deb583
   rollapp.LivenessHeight = TODO // dependent on liveness slashing

   if mustUpgrade {
       lastStateUpdate := k.GetLastStateUpda
       rollapp.FaultyDRS = lastStateUpdate.BDs[last].DrsVersion
   }
}

*/
