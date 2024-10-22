package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// HardFork handles the fraud evidence submitted by the user.
func (k Keeper) HardFork(ctx sdk.Context, rollappID string, fraudHeight uint64) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	err := k.RevertPendingStates(ctx, rollappID, fraudHeight)
	if err != nil {
		return errorsmod.Wrap(err, "revert pending states")
	}

	// update revision number
	// rollapp.RevisionNumber += 1
	// fixme: set liveness height

	rollapp.Frozen = true
	k.SetRollapp(ctx, rollapp)

	// slash the sequencer, clean delayed packets
	err = k.hooks.OnHardFork(ctx, rollappID, fraudHeight)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFraud,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
			sdk.NewAttribute(types.AttributeKeyFraudHeight, fmt.Sprint(fraudHeight)),
		),
	)

	return nil
}

// removes state updates until the one specified and included
func (k Keeper) RevertPendingStates(ctx sdk.Context, rollappID string, fraudHeight uint64) error {
	// find the affected state info index
	// skip if not found (height not committed yet)
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, fraudHeight)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	// check height is not finalized
	if stateInfo.Status == common.Status_FINALIZED {
		return errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", fraudHeight)
	}

	// trunc the state info if needed
	err = k.TruncStateInfo(ctx, stateInfo, fraudHeight)
	if err != nil {
		return errorsmod.Wrap(err, "trunc state info")
	}

	lastIdxToKeep := stateInfo.StateInfoIndex.Index

	// TODO (#631): Prefix store by rollappID for efficient querying
	queuePerHeight := k.GetAllBlockHeightToFinalizationQueue(ctx)
	revertedCount := 0 // Counter for reverted state updates

	for _, queue := range queuePerHeight {
		leftPendingStates := []types.StateInfoIndex{}
		for _, stateInfoIndex := range queue.FinalizationQueue {
			// keep pending packets not related to this rollapp in the queue
			if stateInfoIndex.RollappId != rollappID {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}

			// keep state info indexes with index less than the rollback index
			if stateInfoIndex.Index <= lastIdxToKeep {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}

			// keep pending packets with height less than the rollback height
			k.RemoveStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			revertedCount++ // Increment the counter
		}

		// no change in the queue
		if len(leftPendingStates) == len(queue.FinalizationQueue) {
			continue
		}

		// remove the queue if no pending states left
		if len(leftPendingStates) == 0 {
			k.RemoveBlockHeightToFinalizationQueue(ctx, queue.CreationHeight)
			continue
		}

		// update the queue after removing the reverted states
		k.SetBlockHeightToFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
			CreationHeight:    queue.CreationHeight,
			FinalizationQueue: leftPendingStates,
		})
	}

	ctx.Logger().Info(fmt.Sprintf("Reverted state updates for rollapp: %s, count: %d", rollappID, revertedCount))

	return nil
}

// trunc fraudelent state info
func (k Keeper) TruncStateInfo(ctx sdk.Context, stateInfo *types.StateInfo, fraudHeight uint64) error {
	return nil

	/*
		// heightToKeep := fraudHeight + 1

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
	*/
}
