package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "sequencers-count", SequencersCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sequencer-proposer-bonded", ProposerBondedInvariant(k))
}

func SequencersCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		sequencers := k.AllSequencers(ctx)
		rollapps := k.rollappKeeper.GetAllRollapps(ctx)

		totalCount := 0
		for _, rollapp := range rollapps {
			seqByRollapp := k.RollappSequencers(ctx, rollapp.RollappId)
			bonded := k.RollappSequencersByStatus(ctx, rollapp.RollappId, types.Bonded)
			unbonded := k.RollappSequencersByStatus(ctx, rollapp.RollappId, types.Unbonded)

			if len(seqByRollapp) != len(bonded)+len(unbonded) {
				broken = true
				msg += "sequencer by rollapp length is not equal to sum of bonded, and unbonded " + rollapp.RollappId + "\n"
			}

			totalCount += len(seqByRollapp)
		}

		if totalCount != len(sequencers) {
			broken = true
			msg += "total sequencer count is not equal to sum of sequencers by rollapp\n"
		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencers-count",
			msg,
		), broken
	}
}

// ProposerBondedInvariant checks if the proposer and next proposer are bonded as expected
func ProposerBondedInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.rollappKeeper.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			proposer := k.GetProposer(ctx, rollapp.RollappId)
			if !proposer.Bonded() {
				broken = true
				msg += "proposer is not bonded " + rollapp.RollappId + "\n"
			}
			successor := k.GetSuccessor(ctx, rollapp.RollappId)
			if !successor.Bonded() {
				broken = true
				msg += "successor is not bonded " + rollapp.RollappId + "\n"
			}

		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencer-bonded",
			msg,
		), broken
	}
}
