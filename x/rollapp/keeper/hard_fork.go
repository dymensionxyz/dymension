package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// HardFork handles the fraud evidence submitted by the user.
func (k Keeper) HardFork(ctx sdk.Context, rollappID string, lastValidHeight uint64) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	if !k.ForkAllowed(ctx, rollappID, lastValidHeight) {
		return gerrc.ErrFailedPrecondition.Wrap("fork not allowed")
	}

	lastValidHeight, err := k.RevertPendingStates(ctx, rollappID, lastValidHeight+1)
	if err != nil {
		return errorsmod.Wrap(err, "revert pending states")
	}

	newRevisionHeight := lastValidHeight + 1

	// update revision number
	rollapp.BumpRevision(newRevisionHeight)

	// stop liveness events
	k.ResetLivenessClock(ctx, &rollapp)

	k.SetRollapp(ctx, rollapp)

	// handle the sequencers, clean delayed packets, handle light client
	err = k.hooks.OnHardFork(ctx, rollappID, lastValidHeight)
	if err != nil {
		return errorsmod.Wrap(err, "hard fork callback")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardFork,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
			sdk.NewAttribute(types.AttributeKeyNewRevisionHeight, fmt.Sprint(newRevisionHeight)),
		),
	)

	return nil
}

// RevertPendingStates removes state updates until the one specified and included
// returns the latest height of the state info
func (k Keeper) RevertPendingStates(ctx sdk.Context, rollappID string, newRevisionHeight uint64) (uint64, error) {
	// find the affected state info index
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, newRevisionHeight)
	if err == nil {
		// check the disputed state info is not already finalized
		if stateInfo.Status == common.Status_FINALIZED {
			return 0, errorsmod.Wrapf(types.ErrDisputeAlreadyFinalized, "state info for height %d is already finalized", newRevisionHeight)
		}
	} else if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		// if not found, it's a future height.
		// use latest state info
		s, ok := k.GetLatestStateInfo(ctx, rollappID)
		if !ok {
			return 0, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "no state info found for rollapp: %s", rollappID)
		}
		stateInfo = &s
	} else {
		return 0, err
	}

	// update the last state info before the fraud height
	// it removes all block descriptors after the fraud height
	// and sets the next proposer to the empty string
	stateInfo, err = k.UpdateLastStateInfo(ctx, stateInfo, newRevisionHeight)
	if err != nil {
		return 0, errorsmod.Wrap(err, "update last state info")
	}
	lastStateIdxToKeep := stateInfo.StateInfoIndex.Index

	// clear states updates post the fraud height
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
	err = k.pruneFinalizationsAbove(ctx, rollappID, lastStateIdxToKeep)
	if err != nil {
		return 0, fmt.Errorf("remove finalization queue: %w", err)
	}

	// remove the sequencers heights
	lastStateInfo := k.MustGetStateInfo(ctx, rollappID, lastStateIdxToKeep)
	err = k.PruneSequencerHeights(ctx, mapKeysToSlice(uniqueProposers), lastStateInfo.GetLatestHeight())
	if err != nil {
		return 0, errorsmod.Wrap(err, "prune sequencer heights")
	}

	ctx.Logger().Info(fmt.Sprintf("Reverted state updates for rollapp: %s, count: %d", rollappID, revertedStatesCount))
	return lastStateInfo.GetLatestHeight(), nil
}

// UpdateLastStateInfo truncates the state info to the last valid block before the fraud height.
// It returns the last state
func (k Keeper) UpdateLastStateInfo(ctx sdk.Context, stateInfo *types.StateInfo, fraudHeight uint64) (*types.StateInfo, error) {
	if fraudHeight < stateInfo.StartHeight {
		return nil, errorsmod.Wrapf(gerrc.ErrInternal, "state info start height is greater than fraud height")
	}

	if stateInfo.StartHeight == fraudHeight {
		// If fraud height is at the beginning of the state info, return the previous index to keep
		var ok bool
		*stateInfo, ok = k.GetStateInfo(ctx, stateInfo.StateInfoIndex.RollappId, stateInfo.StateInfoIndex.Index-1)
		if !ok {
			return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "no state info found for rollapp: %s", stateInfo.StateInfoIndex.RollappId)
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
	return stateInfo, nil
}

func (k Keeper) HardForkToLatest(ctx sdk.Context, rollappID string) error {
	lastBatch, ok := k.GetLatestStateInfo(ctx, rollappID)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "no last batch")
	}
	// we invoke a hard fork on the last posted batch without reverting any states
	return k.HardFork(ctx, rollappID, lastBatch.GetLatestHeight())
}

func mapKeysToSlice(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (k Keeper) pruneFinalizationsAbove(ctx sdk.Context, rollappID string, lastStateIdxToKeep uint64) error {
	queuePerHeight, err := k.GetFinalizationQueueByRollapp(ctx, rollappID)
	if err != nil {
		return errorsmod.Wrap(err, "get finalization q by rollapp")
	}
	for _, q := range queuePerHeight {
		leftPendingStates := []types.StateInfoIndex{}
		for _, stateInfoIndex := range q.FinalizationQueue {
			// keep state info indexes with index less than the rollback index
			if stateInfoIndex.Index <= lastStateIdxToKeep {
				leftPendingStates = append(leftPendingStates, stateInfoIndex)
				continue
			}
		}

		if len(leftPendingStates) == 0 {
			if err := k.RemoveFinalizationQueue(ctx, q.CreationHeight, rollappID); err != nil {
				return errorsmod.Wrap(err, "remove finalization queue")
			}
		} else {
			if err := k.SetFinalizationQueue(ctx, types.BlockHeightToFinalizationQueue{
				RollappId:         rollappID,
				CreationHeight:    q.CreationHeight,
				FinalizationQueue: leftPendingStates,
			}); err != nil {
				return errorsmod.Wrap(err, "set finalization queue")
			}
		}
	}
	return nil
}

// is <height,revision> the first of the latest fork?
func (k Keeper) IsFirstHeightOfLatestFork(ctx sdk.Context, rollappId string, revision, height uint64) bool {
	rollapp := k.MustGetRollapp(ctx, rollappId)
	latest := rollapp.LatestRevision().Number
	return rollapp.DidFork() && rollapp.IsRevisionStartHeight(revision, height) && revision == latest
}

// is forking to the latest height going to violate assumptions?
func (k Keeper) ForkLatestAllowed(ctx sdk.Context, rollapp string) bool {
	lastHeight, ok := k.GetLatestHeight(ctx, rollapp)
	if !ok {
		return false
	}
	return k.ForkAllowed(ctx, rollapp, lastHeight)
}

// is the rollback fork going to violate assumptions?
func (k Keeper) ForkAllowed(ctx sdk.Context, rollapp string, lastValidHeight uint64) bool {
	ra := k.MustGetRollapp(ctx, rollapp)
	return 0 < ra.GenesisState.TransferProofHeight && ra.GenesisState.TransferProofHeight <= lastValidHeight
}
