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
)

// HandleFraud handles the fraud evidence submitted by the user.
func (k Keeper) HandleFraud(ctx sdk.Context, rollappID, clientID string, fraudHeight uint64, seqAddr string) error {
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

	// slash the sequencer, clean delayed packets
	err = k.hooks.FraudSubmitted(ctx, rollappID, fraudHeight, seqAddr)
	if err != nil {
		return err
	}

	// freeze the rollapp and revert all pending states
	err = k.FreezeRollapp(ctx, rollappID)
	if err != nil {
		return fmt.Errorf("freeze rollapp: %w", err)
	}

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
