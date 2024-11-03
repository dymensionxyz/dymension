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

	lastCommittedHeight, err := k.RevertPendingStates(ctx, rollappID, fraudHeight)
	if err != nil {
		return errorsmod.Wrap(err, "revert pending states")
	}

	// update revision number
	rollapp.RevisionNumber += 1
	rollapp.RevisionStartHeight = lastCommittedHeight + 1
	k.SetRollapp(ctx, rollapp)

	// handle the sequencers, clean delayed packets, handle light client
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
// returns the latest height of the state info
func (k Keeper) RevertPendingStates(ctx sdk.Context, rollappID string, fraudHeight uint64) (uint64, error) {
	// find the affected state info index
	// skip if not found (fraud height is not committed yet)
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, fraudHeight)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		sinfo, _ := k.GetLatestStateInfo(ctx, rollappID)
		return sinfo.GetLatestHeight(), nil
	} else if err != nil {
		return 0, err
	}

	// check height is not finalized
	if stateInfo.Status == common.Status_FINALIZED {
		return 0, errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", fraudHeight)
	}

	lastStateIdxToKeep := stateInfo.StateInfoIndex.Index
	// trunc the state info if needed
	if stateInfo.StartHeight == fraudHeight {
		// If fraud height is at the beginning of the state info, return the previous index to keep
		lastStateIdxToKeep -= 1
	} else {
		// otherwise, truncate the state info
		k.TruncStateInfo(ctx, stateInfo, fraudHeight)
	}

	// clear state info
	revertedStatesCount := 0 // Counter for reverted state updates
	lastIdx, _ := k.GetLatestStateInfoIndex(ctx, rollappID)
	for i := lastStateIdxToKeep + 1; i <= lastIdx.Index; i++ {
		k.RemoveStateInfo(ctx, rollappID, i)
		revertedStatesCount++ // Increment the counter
	}
	k.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
		RollappId: rollappID,
		Index:     lastStateIdxToKeep,
	})

	// remove all the pending states from the finalization queue
	// we iterate over the queue,
	// - skipping the states that are not related to the rollapp
	// - skipping the states that are less than the rollback index
	queuePerHeight := k.GetAllBlockHeightToFinalizationQueue(ctx) // FIXME (#631): Prefix store by rollappID for efficient querying
	for _, queue := range queuePerHeight {
		leftPendingStates := []types.StateInfoIndex{}
		for _, stateInfoIndex := range queue.FinalizationQueue {
			// keep pending packets not related to this rollapp in the queue
			if stateInfoIndex.RollappId != rollappID {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}

			// keep state info indexes with index less than the rollback index
			if stateInfoIndex.Index <= lastStateIdxToKeep {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}
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

	ctx.Logger().Info(fmt.Sprintf("Reverted state updates for rollapp: %s, count: %d", rollappID, revertedStatesCount))

	return fraudHeight - 1, nil
}

// TruncStateInfo truncates the state info to the last valid block before the fraud height.
// It returns the index of the last state info to keep.
func (k Keeper) TruncStateInfo(ctx sdk.Context, stateInfo *types.StateInfo, fraudHeight uint64) {
	// Remove block descriptors until the one we need to rollback to
	truncatedBDs := stateInfo.BDs.BD[:fraudHeight-stateInfo.StartHeight]

	// Update the state info to reflect truncated data
	stateInfo.NumBlocks = uint64(len(truncatedBDs))
	stateInfo.BDs.BD = truncatedBDs

	// Update the state info in the keeper
	k.SetStateInfo(ctx, *stateInfo)
}

func (k Keeper) HardForkToLatest(ctx sdk.Context, rollappID string) error {
	lastBatch, ok := k.GetLatestStateInfo(ctx, rollappID)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "can't hard fork, no state info found")
	}
	// we invoke a hard fork on the last posted batch without reverting any states
	return k.HardFork(ctx, rollappID, lastBatch.GetLatestHeight()+1)
}
