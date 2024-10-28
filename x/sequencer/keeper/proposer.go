package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) optOutAllSequencers(ctx sdk.Context, rollapp string, excl ...string) {
	seqs := k.RollappSequencers(ctx, rollapp)
	exclMap := make(map[string]struct{}, len(excl))
	for _, addr := range excl {
		exclMap[addr] = struct{}{}
	}
	for _, seq := range seqs {
		if _, ok := exclMap[seq.Address]; !ok {
			seq.OptedIn = false
			k.SetSequencer(ctx, seq)
		}
	}
}

func (k Keeper) ChooseProposer(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded - invariant broken")
		}
		// a valid proposer is already set so there's no need to do anything
		return nil
	}
	seqs := k.GetRollappPotentialProposers(ctx, rollapp)
	proposer = k.proposerChoiceAlgo(ctx, rollapp, seqs)
	k.SetProposer(ctx, rollapp, proposer.Address)
	return nil
}

func (k Keeper) chooseSuccessor(ctx sdk.Context, rollapp string) error {
	successor := k.GetSuccessor(ctx, rollapp)
	if !successor.Sentinel() {
		// a valid successor is already set so there's no need to do anything
		return nil
	}
	proposer := k.GetProposer(ctx, rollapp)
	if proposer.Sentinel() {
		return gerrc.ErrInternal.Wrap("should not choose successor if proposer is sentinel")
	}
	seqs := k.GetRollappPotentialProposers(ctx, rollapp)
	seqs = slices.DeleteFunc(seqs, func(s types.Sequencer) bool { // Not efficient, could optimize.
		return s.Address == proposer.Address
	})
	successor = k.proposerChoiceAlgo(ctx, rollapp, seqs)
	k.SetSuccessor(ctx, rollapp, successor.Address)
	return nil
}

func (k Keeper) isPotentialProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Bonded() && seq.OptedIn
}

func (k Keeper) proposerChoiceAlgo(ctx sdk.Context, rollapp string, seqs []types.Sequencer) types.Sequencer {
	if len(seqs) == 0 {
		return k.SentinelSequencer(ctx)
	}
	// slices package is recommended over sort package
	slices.SortStableFunc(seqs, func(a, b types.Sequencer) int {
		ca := a.TokensCoin()
		cb := b.TokensCoin()
		if ca.IsEqual(cb) {
			return 0
		}
		if ca.IsLT(cb) {
			return -1
		}
		return 1
	})
	return seqs[0]
}

func (k Keeper) IsProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetProposer(ctx, seq.RollappId).Address
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}

// isProposerOrSuccessor returns true if the sequencer requires a notice period before unbonding
// Both the proposer and the next proposer require a notice period
func (k Keeper) isProposerOrSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return k.IsProposer(ctx, seq) || k.IsSuccessor(ctx, seq)
}

// requiresNoticePeriod returns true iff the sequencer requires a notice period before unbonding
func (k Keeper) requiresNoticePeriod(ctx sdk.Context, seq types.Sequencer) bool {
	return k.isProposerOrSuccessor(ctx, seq)
}
