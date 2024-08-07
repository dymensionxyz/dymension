package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "rollapp-state-index", RollappLatestStateIndexInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-count", RollappCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "block-height-to-finalization-queue", BlockHeightToFinalizationQueueInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-by-eip155-key", RollappByEIP155KeyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-finalized-state", RollappFinalizedStateInvariant(k))
	ir.RegisterRoute(types.ModuleName, "liveness-event", LivenessEventInvariant(k))
}

// AllInvariants runs all invariants of the module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := RollappLatestStateIndexInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = RollappCountInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = BlockHeightToFinalizationQueueInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = RollappByEIP155KeyInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = RollappFinalizedStateInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = LivenessEventInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		return "", false
	}
}

// RollappByEIP155KeyInvariant checks that assuming rollapp id is registered with eip155 chain id
// then it should be retrievable by eip155 key
func RollappByEIP155KeyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			rollappID, err := types.NewChainID(rollapp.RollappId)
			if err != nil {
				msg += fmt.Sprintf("rollapp (%s) have invalid rollappId\n", rollapp.RollappId)
				broken = true
				continue
			}

			_, found := k.GetRollappByEIP155(ctx, rollappID.GetEIP155ID())
			if !found {
				msg += fmt.Sprintf("rollapp (%s) have no eip155 key\n", rollapp.RollappId)
				broken = true
				continue
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "rollapp-by-eip155-key",
			msg,
		), broken
	}
}

// BlockHeightToFinalizationQueueInvariant checks that all unfinalized states are in the finalization queue
func BlockHeightToFinalizationQueueInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		for _, rollapp := range k.GetAllRollapps(ctx) {
			if !k.IsRollappStarted(ctx, rollapp.RollappId) {
				continue
			}

			if rollapp.GetFrozen() {
				continue
			}
			latestStateIdx, _ := k.GetLatestStateInfoIndex(ctx, rollapp.RollappId)
			latestFinalizedStateIdx, _ := k.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)

			firstUnfinalizedStateIdx := latestFinalizedStateIdx.Index + 1

			// iterate over all the unfinalized states and make sure they are in the queue
			for i := firstUnfinalizedStateIdx; i <= latestStateIdx.Index; i++ {
				stateInfo, found := k.GetStateInfo(ctx, rollapp.RollappId, i)

				if !found {
					msg += fmt.Sprintf("rollapp (%s) have no stateInfo at index %d\n", rollapp.RollappId, i)
					broken = true
					continue
				}
				creationHeight := stateInfo.CreationHeight
				val, found := k.GetBlockHeightToFinalizationQueue(ctx, creationHeight)
				if !found {
					msg += fmt.Sprintf("finalizationQueue (%d) have no block height\n", creationHeight)
					broken = true
					continue
				}

				// check that our state index is in the queue
				found = false
				for _, idx := range val.FinalizationQueue {
					if idx.RollappId == rollapp.RollappId && idx.Index == i {
						found = true
						break
					}
				}
				if !found {
					msg += fmt.Sprintf("rollapp (%s) have stateInfo at index %d not in the queue\n", rollapp.RollappId, i)
					broken = true
				}
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "block-height-to-finalization-queue",
			msg,
		), broken
	}
}

// RollappCountInvariant checks that the number of rollapps is equal to the number of latestStateInfoIndex
func RollappCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.GetAllRollapps(ctx)
		rollappCount := len(rollapps)
		rollappCountFromIndex := len(k.GetAllLatestStateInfoIndex(ctx))

		if rollappCount == rollappCountFromIndex {
			return "", false
		}

		// If the count is not equal, need to check whether it's due to rollapp that didn't publish state yet
		var noStateRollappCount int
		for _, rollapp := range rollapps {
			if !k.IsRollappStarted(ctx, rollapp.RollappId) {
				noStateRollappCount++
			}
		}

		broken = rollappCount != (rollappCountFromIndex + noStateRollappCount)
		if broken {
			msg = fmt.Sprintf("rollapp count (%d) != latestStateInfoIndex count (%d) + noStateRollapp count (%d)\n", rollappCount, rollappCountFromIndex, noStateRollappCount)
		}

		return sdk.FormatInvariant(
			types.ModuleName, "rollapp-count",
			msg,
		), broken
	}
}

// RollappLatestStateIndexInvariant checks the following invariants per each rollapp that latest state index >= finalized state index
func RollappLatestStateIndexInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			if !k.IsRollappStarted(ctx, rollapp.RollappId) {
				continue
			}

			latestStateIdx, found := k.GetLatestStateInfoIndex(ctx, rollapp.RollappId)
			if !found {
				msg += fmt.Sprintf("rollapp (%s) have no latestStateIdx\n", rollapp.RollappId)
				broken = true
				break
			}

			latestFinalizedStateIdx, _ := k.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
			// not found is ok, it means no finalized state yet

			if latestStateIdx.Index < latestFinalizedStateIdx.Index {
				msg += fmt.Sprintf("rollapp (%s) have latestStateIdx < latestFinalizedStateIdx\n", rollapp.RollappId)
				broken = true
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "rollapp-state-index",
			msg,
		), broken
	}
}

// RollappFinalizedStateInvariant checks that all the states until latest finalized state are finalized
func RollappFinalizedStateInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			// If the rollapp is not started, we don't need to check
			if !k.IsRollappStarted(ctx, rollapp.RollappId) {
				continue
			}

			// If we didn't finalize any state yet, we don't need to check
			latestFinalizedStateIdx, found := k.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)
			if !found {
				continue
			}

			for i := uint64(1); i <= latestFinalizedStateIdx.Index; i++ {
				stateInfo, found := k.GetStateInfo(ctx, rollapp.RollappId, i)
				if !found {
					msg += fmt.Sprintf("rollapp (%s) have no stateInfo at index %d\n", rollapp.RollappId, i)
					broken = true
				}

				if stateInfo.Status != commontypes.Status_FINALIZED {
					msg += fmt.Sprintf("rollapp (%s) have stateInfo at index %d not finalized\n", rollapp.RollappId, i)
					broken = true
				}
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "rollapp-finalized-state",
			msg,
		), broken
	}
}

// LivenessEventInvariant checks for all rollapps that the liveness event height, if any, is accurate,
// in that there is actually an event stored at that height. Moreover, there should not be any events
// stored which don't correspond to a liveness event height.
func LivenessEventInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		rollapps := k.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			if !ra.LastStateUpdateHeightIsSet() {
				continue
			}
			events := k.GetLivenessEvents(ctx, &ra.LivenessEventHeight)
			cnt := 0
			for _, event := range events {
				if event.RollappId == ra.RollappId {
					cnt++
				}
			}
			if cnt != 1 {
				broken = true
				msg += fmt.Sprintf("| rollapp stored event but wrong number found in queue: rollapp: %s: event height: %d: found: %d", ra.RollappId, ra.LivenessEventHeight, cnt)
			}
		}
		for _, e := range k.GetLivenessEvents(ctx, nil) {
			ra, ok := k.GetRollapp(ctx, e.RollappId)
			if !ok {
				broken = true
				msg += fmt.Sprintf("| event stored but rollapp not found: rollapp id: %s", e.RollappId)
				continue
			}
			if ra.LivenessEventHeight != e.HubHeight {
				broken = true
				msg += fmt.Sprintf("| event stored but rollapp has a different liveness event height: rollapp: %s"+
					", height stored on rollapp: %d: height on event: %d", e.RollappId, ra.LivenessEventHeight, e.HubHeight,
				)
			}

		}

		return sdk.FormatInvariant(
			types.ModuleName, "liveness-event",
			msg,
		), broken
	}
}
