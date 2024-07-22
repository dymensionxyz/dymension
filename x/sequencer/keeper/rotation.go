package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// MatureSequencersWithNoticePeriod moves all the sequencers that have finished their notice period
func (k Keeper) MatureSequencersWithNoticePeriod(ctx sdk.Context, currTime time.Time) {
	seqs := k.GetMatureNoticePeriodSequencers(ctx, currTime)
	for _, seq := range seqs {
		k.StartRotation(ctx, seq.RollappId)
	}
}

// IsRotating returns true if the rollapp is currently rotating proposers
func (k Keeper) IsRotating(ctx sdk.Context, rollappId string) bool {
	return k.HasNextProposer(ctx, rollappId)
}

// IsNoticePeriodRequired returns true if the sequencer requires a notice period before unbonding
func (k Keeper) IsNoticePeriodRequired(ctx sdk.Context, seq types.Sequencer) bool {
	return k.IsProposer(ctx, seq.RollappId, seq.SequencerAddress) || k.IsNextProposer(ctx, seq.RollappId, seq.SequencerAddress)
}

// ExpectedNextProposer returns the next proposer for a rollapp
func (k Keeper) ExpectedNextProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	if len(seqs) == 0 {
		return types.Sequencer{}
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].Tokens.IsAllGT(seqs[j].Tokens)
	})

	// filter out proposer and nextProposer
	active, _ := k.GetActiveSequencer(ctx, rollappId)
	next, _ := k.GetNextProposer(ctx, rollappId)
	for _, s := range seqs {
		if s.SequencerAddress == active.SequencerAddress || s.SequencerAddress == next.SequencerAddress {
			continue
		}
		return s
	}

	return types.Sequencer{}
}

// StartRotation sets the nextSequencer for the rollapp.
// This function will not clear the current proposer
// This function called when the sequencer has finished its notice period
func (k Keeper) StartRotation(ctx sdk.Context, rollappId string) {
	// next proposer can be empty if there are no bonded sequencers available
	nextProposer := k.ExpectedNextProposer(ctx, rollappId)
	k.SetNextProposer(ctx, rollappId, nextProposer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotationStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeyNextProposer, nextProposer.SequencerAddress),
		),
	)
}

// RotateProposer completes the sequencer rotation flow.
// it will start unbonding the current proposer, and set new proposer from the bonded sequencers
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) error {
	proposer, ok := k.GetActiveSequencer(ctx, rollappId)
	if ok {
		_, err := k.setSequencerToUnbonding(ctx, &proposer)
		if err != nil {
			return err
		}
	}

	nextProposer, _ := k.GetNextProposer(ctx, rollappId) // nextProposer is guaranteed to exist by caller
	k.SetActiveSequencer(ctx, rollappId, nextProposer)
	k.RemoveNextProposer(ctx, rollappId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposer.SequencerAddress),
		),
	)

	return nil
}
