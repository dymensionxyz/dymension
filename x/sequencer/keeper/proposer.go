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

func (k Keeper) RollappPotentialProposers(ctx sdk.Context, rollappId string) []types.Sequencer {
	seqs := k.RollappBondedSequencers(ctx, rollappId)
	seqs = slices.DeleteFunc(seqs, func(seq types.Sequencer) bool {
		return !seq.IsPotentialProposer()
	})
	return append(seqs, k.SentinelSequencer(ctx))
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
		proposer, err := ProposerChoiceAlgo(seqs)
		if err != nil {
			return err
		}
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

// ChooseSuccessor will assign a successor. It won't replace an existing one.
// It will prioritise non sentinel
func (k Keeper) chooseSuccessor(ctx sdk.Context, rollapp string) error {
	successor := k.GetSuccessor(ctx, rollapp)
	if !successor.Sentinel() {
		// a valid successor is already set so there's no need to do anything
		return nil
	}
	proposer := k.GetProposer(ctx, rollapp)
	if proposer.Sentinel() {
		return nil
	}
	seqs := k.RollappPotentialProposers(ctx, rollapp)
	successor, err := ProposerChoiceAlgo(seqs)
	if err != nil {
		return err
	}
	k.SetSuccessor(ctx, rollapp, successor.Address)
	return nil
}

// ProposerChoiceAlgo : choose the one with most bond
// Requires sentinel to be passed in, as last resort.
func ProposerChoiceAlgo(seqs []types.Sequencer) (types.Sequencer, error) {
	if len(seqs) == 0 {
		return types.Sequencer{}, gerrc.ErrInternal.Wrap("seqs must at least include sentinel")
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
	return seqs[0], nil
}

func (k Keeper) IsProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetProposer(ctx, seq.RollappId).Address
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}
