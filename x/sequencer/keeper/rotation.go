package keeper

import (
	"sort"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) time.Time {
	completionTime := ctx.BlockTime().Add(k.NoticePeriod(ctx))
	seq.NoticePeriodTime = completionTime

	k.UpdateSequencer(ctx, seq)
	k.AddSequencerToNoticePeriodQueue(ctx, seq)

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
// In the middle of rotation, the next proposer required a notice period as well.
func (k Keeper) MatureSequencersWithNoticePeriod(ctx sdk.Context, currTime time.Time) {
	seqs := k.GetMatureNoticePeriodSequencers(ctx, currTime)
	for _, seq := range seqs {
		if k.isProposer(ctx, seq.RollappId, seq.Address) {
			k.startRotation(ctx, seq.RollappId)
			k.removeNoticePeriodSequencer(ctx, seq)
		}
		// next proposer cannot mature it's notice period until the current proposer has finished rotation
		// minor effect as notice_period >>> rotation time
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
// it selects the next proposer from the bonded sequencers by bond amount
// if there are no bonded sequencers, it returns an empty sequencer
func (k Keeper) ExpectedNextProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	// if nextProposer is set, were in the middle of rotation. The expected next proposer cannot change
	seq, ok := k.GetNextProposer(ctx, rollappId)
	if ok {
		return seq
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
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

// CompleteRotation completes the sequencer rotation flow.
// It's called when a last state update is received from the active, rotating sequencer.
// it will start unbonding the current proposer, and sets the nextProposer as the proposer.
func (k Keeper) CompleteRotation(ctx sdk.Context, rollappId string) error {
	proposer, ok := k.GetProposer(ctx, rollappId)
	if !ok {
		return errors.Wrapf(gerrc.ErrInternal, "proposer not set for rollapp %s", rollappId)
	}
	nextProposer, ok := k.GetNextProposer(ctx, rollappId)
	if !ok {
		return errors.Wrapf(gerrc.ErrInternal, "next proposer not set for rollapp %s", rollappId)
	}

	// start unbonding the current proposer
	k.startUnbondingPeriodForSequencer(ctx, &proposer)

	// change the proposer
	k.removeNextProposer(ctx, rollappId)
	k.SetProposer(ctx, rollappId, nextProposer.Address)

	if nextProposer.Address == NO_SEQUENCER_AVAILABLE {
		k.Logger(ctx).Info("Rollapp left with no proposer.", "RollappID", rollappId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposer.Address),
		),
	)

	return nil
}
