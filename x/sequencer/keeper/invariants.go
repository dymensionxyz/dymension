package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "sequencers-count", SequencersCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sequencer-proposer-bonded", ProposerBondedInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sequencer-min-bond", SequencerMinBondInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sequencer-positive-balance-post-bond-reduction", SequencerPositiveBalancePostBondReduction(k))
}

// AllInvariants runs all invariants of the x/sequencer module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := SequencersCountInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = ProposerBondedInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return "", false
	}
}

func SequencersCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		sequencers := k.GetAllSequencers(ctx)
		rollapps := k.rollappKeeper.GetAllRollapps(ctx)

		totalCount := 0
		for _, rollapp := range rollapps {
			seqByRollapp := k.GetSequencersByRollapp(ctx, rollapp.RollappId)
			bonded := k.GetSequencersByRollappByStatus(ctx, rollapp.RollappId, types.Bonded)
			unbonding := k.GetSequencersByRollappByStatus(ctx, rollapp.RollappId, types.Unbonding)
			unbonded := k.GetSequencersByRollappByStatus(ctx, rollapp.RollappId, types.Unbonded)

			if len(seqByRollapp) != len(bonded)+len(unbonding)+len(unbonded) {
				broken = true
				msg += "sequencer by rollapp length is not equal to sum of bonded, unbonding and unbonded " + rollapp.RollappId + "\n"
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
			active, ok := k.GetProposer(ctx, rollapp.RollappId)
			if ok && active.Status != types.Bonded {
				broken = true
				msg += "active sequencer is not bonded " + rollapp.RollappId + "\n"
			}

			next := k.ExpectedNextProposer(ctx, rollapp.RollappId)
			if !next.IsEmpty() && next.Status != types.Bonded {
				broken = true
				msg += "next sequencer is not bonded " + rollapp.RollappId + "\n"
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencer-bonded",
			msg,
		), broken
	}
}

// SequencerMinBondInvariant checks if the sequencer always maintains the minimum bond as long as it is bonded status
func SequencerMinBondInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		minBond := sdk.NewCoins(k.MinBond(ctx))
		sequencers := k.GetAllSequencers(ctx)
		for _, seq := range sequencers {
			if seq.Status == types.Bonded && !seq.Tokens.IsAllGTE(minBond) {
				broken = true
				msg += "bonded sequencer does not have minimum bond " + seq.Address + "\n"
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencer-min-bond",
			msg,
		), broken
	}
}

// SequencerPositiveBalancePostBondReduction checks if the sequencer maintains a positive balance after all bond reductions are applied
func SequencerPositiveBalancePostBondReduction(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		sequencers := k.GetAllSequencers(ctx)
		for _, seq := range sequencers {
			effectiveBond := seq.Tokens
			if bondReductions := k.getSequencerDecreasingBonds(ctx, seq.Address); len(bondReductions) > 0 {
				for _, bd := range bondReductions {
					effectiveBond = effectiveBond.Sub(bd.DecreaseBondAmount)
				}
			}
			if effectiveBond.IsAnyNegative() {
				broken = true
				msg += "sequencer will have negative balance after bond reduction " + seq.Address + "\n"
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencer-positive-balance-post-bond-reduction",
			msg,
		), broken
	}
}
