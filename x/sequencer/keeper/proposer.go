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

// RecoverFromSentinel will assign a new proposer to the rollapp.
// It will choose a new proposer from the list of potential proposers.
// The rollapp must be halted and with potential proposer available.
func (k Keeper) RecoverFromSentinel(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)

	if !proposer.Sentinel() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "proposer is not sentinel")
	}

	successor, err := ProposerChoiceAlgo(k.RollappPotentialProposers(ctx, rollapp))
	if err != nil {
		return err
	}
	if successor.Sentinel() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no valid proposer found")
	}
	k.SetProposer(ctx, rollapp, successor.Address)

	err = k.hooks.AfterSetRealProposer(ctx, rollapp, successor)
	if err != nil {
		return errorsmod.Wrap(err, "recovery from halt callbacks")
	}
	if err := uevent.EmitTypedEvent(ctx, &types.EventProposerChange{
		Rollapp: rollapp,
		Before:  types.SentinelSeqAddr,
		After:   successor.Address,
	}); err != nil {
		return err
	}

	return nil
}

func (k Keeper) IsProposer(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetProposer(ctx, seq.RollappId).Address
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}
