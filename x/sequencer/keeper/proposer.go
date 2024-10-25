package keeper

import (
	"slices"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) chooseProposer(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded - invariant broken")
		}
		return nil
	}
	successor := k.GetSuccessor(ctx, rollapp)
	k.SetProposer(ctx, rollapp, successor.Address)
	k.SetSuccessor(ctx, rollapp, types.SentinelSeqAddr)
	if k.GetProposer(ctx, rollapp).Sentinel() {
		seqs := k.GetRollappPotentialProposers(ctx, rollapp)
		proposer := k.proposerChoiceAlgo(ctx, rollapp, seqs)
		k.SetProposer(ctx, rollapp, proposer.Address)
	}
	return nil
}

func (k Keeper) chooseSuccessor(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	if proposer.Sentinel() {
		return gerrc.ErrInternal.Wrap("can not choose successor if proposer is sentinel")
	}
	successor := k.GetSuccessor(ctx, rollapp)
	if successor.Sentinel() {
		seqs := k.GetRollappPotentialProposers(ctx, rollapp)
		seqs = slices.DeleteFunc(seqs, func(s types.Sequencer) bool { // Not efficient, could optimize.
			return s.Address == proposer.Address
		})
		successor := k.proposerChoiceAlgo(ctx, rollapp, seqs)
		k.SetSuccessor(ctx, rollapp, successor.Address)

	}
	return nil
}

func (k Keeper) proposerChoiceAlgo(ctx sdk.Context, rollapp string, seqs []types.Sequencer) types.Sequencer {
	if len(seqs) == 0 {
		return k.SentinelSequencer(ctx)
	}
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].TokensCoin().IsGTE(seqs[j].TokensCoin())
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
