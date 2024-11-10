package keeper

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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

// RecoverFromSentinelProposerIfNeeded will assign a proposer to the rollapp. It won't replace the incumbent proposer
// if they are not sentinel. Otherwise it will prioritize a non sentinel successor. Finally, it
// choose one based on an algorithm.
// The result can be the sentinel sequencer.
func (k Keeper) RecoverFromSentinelProposerIfNeeded(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)

	// a valid proposer is already set so there's no need to do anything
	if !proposer.Sentinel() {
		return nil
	}

	before := proposer
	successor := k.GetSuccessor(ctx, rollapp)
	// if successor is sentinel, we attempt to find a non sentinel successor
	if successor.Sentinel() {
		seqs := k.RollappPotentialProposers(ctx, rollapp)
		next, err := ProposerChoiceAlgo(seqs)
		if err != nil {
			return err
		}
		successor = next
	}

	k.SetProposer(ctx, rollapp, successor.Address)
	k.SetSuccessor(ctx, rollapp, types.SentinelSeqAddr)

	if !successor.Sentinel() {
		k.hooks.AfterRecoveryFromHalt(ctx, rollapp, before, successor)

		if err := uevent.EmitTypedEvent(ctx, &types.EventProposerChange{
			Rollapp: rollapp,
			Before:  before.Address,
			After:   successor.Address,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) IsProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetProposer(ctx, seq.RollappId).Address
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}
