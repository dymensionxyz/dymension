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
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
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
		if k.isProposer(ctx, seq.RollappId, seq.Address) {
			k.startRotation(ctx, seq.RollappId)
			k.removeNoticePeriodSequencer(ctx, seq)
		}
	}
}

// IsRotating returns true if the rollapp is currently in the process of rotation.
// A process of rotation is defined by the existence of a next proposer. The next proposer can also be a "dummy" sequencer (i.e empty) in case no sequencer came. This is still considered rotation
// as the sequencer is rotating to an empty one (i.e gracefully leaving the rollapp).
// The next proposer can only be set after the notice period is over. The rotation period is over after the proposer sends his last batch.
func (k Keeper) IsRotating(ctx sdk.Context, rollappId string) bool {
	return k.isNextProposerSet(ctx, rollappId)
}

// isNoticePeriodRequired returns true if the sequencer requires a notice period before unbonding
// Both the proposer and the next proposer require a notice period
func (k Keeper) isNoticePeriodRequired(ctx sdk.Context, seq types.Sequencer) bool {
	return k.isProposer(ctx, seq.RollappId, seq.Address) || k.isNextProposer(ctx, seq.RollappId, seq.Address)
}

// ExpectedNextProposer returns the next proposer for a rollapp
func (k Keeper) ExpectedNextProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	// if nextProposer is set, were in the middle of rotation
	seqAddr, ok := k.GetNextProposerAddr(ctx, rollappId)
	if ok {
		return k.MustGetSequencer(ctx, seqAddr)
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
		if s.Address != proposer.Address {
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
	k.setNextProposer(ctx, rollappId, nextProposer.Address)

	k.Logger(ctx).Info("rotation started", "rollappId", rollappId, "nextProposer", nextProposer.Address)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotationStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeyNextProposer, nextProposer.Address),
		),
	)
}

// RotateProposer completes the sequencer rotation flow.
// It's called when a last state update is received from the active, rotating sequencer.
// it will start unbonding the current proposer, and set new proposer from the bonded sequencers
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) {
	nextProposerAddr, ok := k.GetNextProposerAddr(ctx, rollappId)
	if !ok { // nextProposer is guaranteed to be set by caller
		k.Logger(ctx).Error("next proposer not set. rotation didn't completed", "rollappId", rollappId)
		return
	}

	// start unbonding the current proposer
	proposer, ok := k.GetProposer(ctx, rollappId)
	if ok {
		k.startUnbondingPeriodForSequencer(ctx, &proposer)
	}

	k.removeNextProposer(ctx, rollappId)
	k.SetProposer(ctx, rollappId, nextProposerAddr)

	if nextProposerAddr == NO_SEQUENCER_AVAILABLE {
		k.Logger(ctx).Info("Rollapp left with no proposer.", "RollappID", rollappId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposerAddr),
		),
	)
}
