package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type Invar = func(sdk.Context) (error, bool)
type InvarNamed struct {
	name  string
	invar func(Keeper) Invar
}

var invars = []InvarNamed{
	{"sequencers-count", SequencersCountInvariant},
	{"sequencers-proposer-bonded", ProposerBondedInvariant},
}

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	for _, invar := range invars {
		ir.RegisterRoute(types.ModuleName, invar.name, func(ctx sdk.Context) (string, bool) {
			err, broken := invar.invar(k)(ctx)
			return err.Error(), broken
		})
	}
}

func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, invar := range invars {
			err, broken := invar.invar(k)(ctx)
			if broken {
				return err.Error(), broken
			}
		}
	}
}

func SequencersCountInvariant(k Keeper) Invar {
	return func(ctx sdk.Context) (error, bool) {
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

		return errors.New(msg), broken
	}
}

// ProposerBondedInvariant checks if the proposer and next proposer are bonded as expected
func ProposerBondedInvariant(k Keeper) Invar {
	return func(ctx sdk.Context) (error, bool) {
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
		return errors.New(msg), broken
	}
}
