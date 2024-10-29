package keeper

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// OptOutAllSequencers : change every sequencer of the rollapp to be opted out.
// Can optionally pass a list of exclusions: those sequencers won't be modified.
func (k Keeper) optOutAllSequencers(ctx sdk.Context, rollapp string, excl ...string) error {
	seqs := k.RollappSequencers(ctx, rollapp)
	exclMap := make(map[string]struct{}, len(excl))
	for _, addr := range excl {
		exclMap[addr] = struct{}{}
	}
	for _, seq := range seqs {
		if _, ok := exclMap[seq.Address]; !ok {
			if err := seq.SetOptedIn(ctx, false); err != nil {
				return errorsmod.Wrap(err, "set opted in")
			}
			k.SetSequencer(ctx, seq)
		}
	}
	return nil
}

// ChooseProposer will assign a proposer to the rollapp. It won't replace the incumbent proposer
// if they are not sentinel. Otherwise it will prioritise a non sentinel successor. Finally, it
// choose one based on an algorithm.
// The result can be the sentinel sequencer.
func (k Keeper) ChooseProposer(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	before := proposer

	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded - invariant broken")
		}
		// a valid proposer is already set so there's no need to do anything
		return nil
	}
	successor := k.GetSuccessor(ctx, rollapp)
	k.SetProposer(ctx, rollapp, successor.Address)
	k.SetSuccessor(ctx, rollapp, types.SentinelSeqAddr)
	if k.GetProposer(ctx, rollapp).Sentinel() {
		seqs := k.RollappPotentialProposers(ctx, rollapp)
		proposer := ProposerChoiceAlgo(seqs)
		k.SetProposer(ctx, rollapp, proposer.Address)
	}

	after := k.GetProposer(ctx, rollapp)
	if before.Address != after.Address {
		k.hooks.AfterChooseNewProposer(ctx, rollapp, before, after)

		if err := uevent.EmitTypedEvent(ctx, &types.EventProposerChange{
			Rollapp: rollapp,
			Before:  before.Address,
			After:   after.Address,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ChooseSuccesor will assign a successor. It won't replace an existing one.
// It will prioritise non sentinel
func (k Keeper) chooseSuccessor(ctx sdk.Context, rollapp string) {
	successor := k.GetSuccessor(ctx, rollapp)
	if !successor.Sentinel() {
		// a valid successor is already set so there's no need to do anything
		// TODO: a necessary check?
		return
	}
	proposer := k.GetProposer(ctx, rollapp)
	if proposer.Sentinel() {
		return
	}
	seqs := k.RollappPotentialProposers(ctx, rollapp)
	successor = ProposerChoiceAlgo(seqs)
	k.SetSuccessor(ctx, rollapp, successor.Address)
	return
}

// isPotentialProposer says if a sequencer can potentially be allowed to propose
// note: will be true for sentinel
func (k Keeper) isPotentialProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Bonded() && seq.OptedIn
}

// ProposerChoiceAlgo : choose the one with most bond
// Requires sentinel to be passed in, as last resort.
func ProposerChoiceAlgo(seqs []types.Sequencer) types.Sequencer {
	if len(seqs) == 0 {
		panic("seqs must at least include sentinel")
	}
	// slices package is recommended over sort package
	slices.SortStableFunc(seqs, func(a, b types.Sequencer) int {
		ca := a.TokensCoin()
		cb := b.TokensCoin()
		if ca.IsEqual(cb) {
			return 0
		}

		// flipped to sort decreasing
		if ca.IsLT(cb) {
			return 1
		}
		return -1
	})
	return seqs[0]
}

func (k Keeper) IsProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetProposer(ctx, seq.RollappId).Address
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}

func (k Keeper) isProposerOrSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return k.IsProposer(ctx, seq) || k.IsSuccessor(ctx, seq)
}

// requiresNoticePeriod returns true iff the sequencer requires a notice period before unbonding
func (k Keeper) requiresNoticePeriod(ctx sdk.Context, seq types.Sequencer) bool {
	return k.isProposerOrSuccessor(ctx, seq)
}
