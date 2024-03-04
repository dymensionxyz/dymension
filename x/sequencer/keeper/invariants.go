package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "sequencers-count", SequencersCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sequencers-per-rollapp", SequencersPerRollappInvariant(k))
}

// AllInvariants runs all invariants of the x/sequencer module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := SequencersCountInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = SequencersPerRollappInvariant(k)(ctx)
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

func SequencersPerRollappInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		rollapps := k.rollappKeeper.GetAllRollapps(ctx)
		for _, rollapp := range rollapps {
			bonded := k.GetSequencersByRollappByStatus(ctx, rollapp.RollappId, types.Bonded)
			unbonding := k.GetSequencersByRollappByStatus(ctx, rollapp.RollappId, types.Unbonding)

			// rollapp.MaxSequencers
			if len(bonded)+len(unbonding) > int(rollapp.MaxSequencers) {
				broken = true
				msg += "too many sequencers for rollapp " + rollapp.RollappId + "\n"
			}

			if len(bonded) == 0 {
				continue
			}

			proposerFound := false
			for _, seq := range bonded {
				if seq.Proposer {
					if proposerFound {
						broken = true
						msg += "more than one proposer for rollapp " + rollapp.RollappId + "\n"
					}
					proposerFound = true
				}
			}
			if !proposerFound {
				broken = true
				msg += "no proposer for rollapp " + rollapp.RollappId + "\n"
			}

		}

		return sdk.FormatInvariant(
			types.ModuleName, "sequencers-per-rollapp",
			msg,
		), broken
	}
}
