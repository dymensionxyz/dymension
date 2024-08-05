package keeper

import (
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) time.Time {
	completionTime := ctx.BlockTime().Add(k.NoticePeriod(ctx))
	seq.NoticePeriodTime = completionTime

	k.UpdateSequencer(ctx, *seq)
	k.AddSequencerToNoticePeriodQueue(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, seq.SequencerAddress),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.String()),
		),
	)

	return completionTime
}

// MatureSequencersWithNoticePeriod start rotation flow for all sequencers that have finished their notice period
// The next proposer is set to the next bonded sequencer
// The hub will expect a "last state update" from the sequencer to start unbonding
func (k Keeper) MatureSequencersWithNoticePeriod(ctx sdk.Context, currTime time.Time) {
	seqs := k.GetMatureNoticePeriodSequencers(ctx, currTime)
	for _, seq := range seqs {
		if k.isProposer(ctx, seq.RollappId, seq.SequencerAddress) {
			k.startRotation(ctx, seq.RollappId)
			k.removeNoticePeriodSequencer(ctx, seq)
		}
	}
}

// IsRotating returns true if the rollapp is currently in the process of rotation.
// A process of rotation is defined by the time between when the proposer notice period is over till the proposer sends his last batch.
func (k Keeper) IsRotating(ctx sdk.Context, rollappId string) bool {
	return k.hasNextProposer(ctx, rollappId)
}

// isNoticePeriodRequired returns true if the sequencer requires a notice period before unbonding
// Both the proposer and the next proposer require a notice period
func (k Keeper) isNoticePeriodRequired(ctx sdk.Context, seq types.Sequencer) bool {
	return k.isProposer(ctx, seq.RollappId, seq.SequencerAddress) || k.isNextProposer(ctx, seq.RollappId, seq.SequencerAddress)
}

// ExpectedNextProposer returns the next proposer for a rollapp
func (k Keeper) ExpectedNextProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	// if nextProposer is set, were in the middle of rotation
	seq, ok := k.GetNextProposer(ctx, rollappId)
	if ok {
		return seq
	}

	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	if len(seqs) == 0 {
		return types.Sequencer{}
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].Tokens.IsAllGT(seqs[j].Tokens)
	})

	// return the first sequencer that is not the proposer
	proposer, _ := k.GetProposer(ctx, rollappId)
	for _, s := range seqs {
		if s.SequencerAddress != proposer.SequencerAddress {
			return s
		}
	}

	return types.Sequencer{}
}

// startRotation sets the nextSequencer for the rollapp.
// This function will not clear the current proposer
// This function called when the sequencer has finished its notice period
func (k Keeper) startRotation(ctx sdk.Context, rollappId string) {
	// next proposer can be empty if there are no bonded sequencers available
	nextProposer := k.ExpectedNextProposer(ctx, rollappId)
	k.setNextProposer(ctx, rollappId, nextProposer.SequencerAddress)

	k.Logger(ctx).Info("rotation started", "rollappId", rollappId, "nextProposer", nextProposer.SequencerAddress)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotationStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeyNextProposer, nextProposer.SequencerAddress),
		),
	)
}

// RotateProposer completes the sequencer rotation flow.
// It's called when a last state update is received from the active, rotating sequencer.
// it will start unbonding the current proposer, and set new proposer from the bonded sequencers
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) {
	// start unbonding the current proposer
	proposer, ok := k.GetProposer(ctx, rollappId)
	if ok {
		k.startUnbondingPeriodForSequencer(ctx, &proposer)
	}

	nextProposer, _ := k.GetNextProposer(ctx, rollappId) // nextProposer is guaranteed to exist by caller
	k.removeNextProposer(ctx, rollappId)
	k.SetProposer(ctx, rollappId, nextProposer.SequencerAddress)

	if nextProposer.SequencerAddress == "" {
		k.Logger(ctx).Info("Rollapp left with no proposer.", "RollappID", rollappId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposer.SequencerAddress),
		),
	)
}
