package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "rollapp-state-index", RollappLatestStateIndexInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-count", RollappCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "block-height-to-finalization-queue", BlockHeightToFinalizationQueueInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rollapp-by-eip155-key", RollappByEIP155KeyInvariant(k))
}

// AllInvariants runs all invariants of the X/bank module.
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
		return "", false
	}
}

func RollappByEIP155KeyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			eip155, err := types.ParseChainID(rollapp.RollappId)
			if err != nil {
				msg += fmt.Sprintf("rollapp (%s) have invalid rollappId\n", rollapp.RollappId)
				broken = true
				continue
			}

			// not breaking invariant, as eip155 format is not required
			if eip155 == nil {
				continue
			}

			_, found := k.GetRollappByEIP155(ctx, eip155.Uint64())
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
			latestStateIdx, _ := k.GetLatestStateInfoIndex(ctx, rollapp.RollappId)
			latestFinalizedStateIdx, _ := k.GetLatestFinalizedStateIndex(ctx, rollapp.RollappId)

			firstUnfinalizedStateIdx := latestFinalizedStateIdx.Index + 1

			// iterate over all the unfinalzied states and make sure they are in the queue
			for i := firstUnfinalizedStateIdx; i <= latestStateIdx.Index; i++ {
				stateInfo, found := k.GetStateInfo(ctx, rollapp.RollappId, i)
				if !found {
					msg += fmt.Sprintf("rollapp (%s) have no stateInfo at index %d\n", rollapp.RollappId, i)
					broken = true
					continue
				}
				creationHeight := stateInfo.CreationHeight
				// finalizationTime := creationHeight + k.DisputePeriodInBlocks(ctx)

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

		// If the count is not equal, need to check wether it's due to rollapp that didn't publish state yet
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
