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
	err = k.hooks.OnHardFork(ctx, rollappID, lastCommittedHeight+1)
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
		s, ok := k.GetLatestStateInfo(ctx, rollappID)
		if !ok {
			return 0, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "no state info found for rollapp: %s", rollappID)
		}
		stateInfo = &s
	} else if err != nil {
		return 0, err
	}

	// check height is not finalized
	if stateInfo.Status == common.Status_FINALIZED {
		return 0, errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", fraudHeight)
	}

	// update the last state info before the fraud height
	// it removes all block descriptors after the fraud height
	// and sets the next proposer to the empty string
	lastStateIdxToKeep, err := k.UpdateLastStateInfo(ctx, stateInfo, fraudHeight)
	if err != nil {
		return 0, errorsmod.Wrap(err, "update last state info")
	}

	// clear pending states post the fraud height
	revertedStatesCount := 0                     // Counter for reverted state updates
	uniqueProposers := make(map[string]struct{}) // Map to manage unique proposers

	lastIdx, _ := k.GetLatestStateInfoIndex(ctx, rollappID)
	for i := lastStateIdxToKeep + 1; i <= lastIdx.Index; i++ {
		// Add the proposer to the unique map
		uniqueProposers[k.MustGetStateInfo(ctx, rollappID, i).Sequencer] = struct{}{}

		// clear the state info
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

	// remove the sequencer heights
	lastStateInfo := k.MustGetStateInfo(ctx, rollappID, lastStateIdxToKeep)
	err = k.PruneSequencerHeights(ctx, mapKeysToSlice(uniqueProposers), lastStateInfo.GetLatestHeight())
	if err != nil {
		return 0, errorsmod.Wrap(err, "prune sequencer heights")
	}

	ctx.Logger().Info(fmt.Sprintf("Reverted state updates for rollapp: %s, count: %d", rollappID, revertedStatesCount))
	return lastStateInfo.GetLatestHeight(), nil
}

// UpdateLastStateInfo truncates the state info to the last valid block before the fraud height.
// It returns the index of the last state info to keep.
func (k Keeper) UpdateLastStateInfo(ctx sdk.Context, stateInfo *types.StateInfo, fraudHeight uint64) (uint64, error) {
	if fraudHeight < stateInfo.StartHeight {
		return 0, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "state info start height is greater than fraud height")
	}

	if stateInfo.StartHeight == fraudHeight {
		// If fraud height is at the beginning of the state info, return the previous index to keep
		var ok bool
		*stateInfo, ok = k.GetStateInfo(ctx, stateInfo.StateInfoIndex.RollappId, stateInfo.StateInfoIndex.Index-1)
		if !ok {
			return 0, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "no state info found for rollapp: %s", stateInfo.StateInfoIndex.RollappId)
		}
	} else if stateInfo.GetLatestHeight() >= fraudHeight {
		// Remove block descriptors until the one we need to rollback to
		truncatedBDs := stateInfo.BDs.BD[:fraudHeight-stateInfo.StartHeight]

		// Update the state info to reflect truncated data
		stateInfo.NumBlocks = uint64(len(truncatedBDs))
		stateInfo.BDs.BD = truncatedBDs
	}

	// Update the state info in the keeper
	stateInfo.NextProposer = ""
	k.SetStateInfo(ctx, *stateInfo)
	return stateInfo.StateInfoIndex.Index, nil
}

func (k Keeper) HardForkToLatest(ctx sdk.Context, rollappID string) error {
	lastBatch, ok := k.GetLatestStateInfo(ctx, rollappID)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "can't hard fork, no state info found")
	}
	// we invoke a hard fork on the last posted batch without reverting any states
	return k.HardFork(ctx, rollappID, lastBatch.GetLatestHeight()+1)
}

func mapKeysToSlice(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
