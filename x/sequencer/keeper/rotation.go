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
		// set the next proposer for the rollapp
		k.StartRotation(ctx, seq.RollappId)
	}
}

// TODO: refactor to use store for optimization
func (k Keeper) GetRollappProposer(ctx sdk.Context, rollappId string) (seq *types.Sequencer) {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	for _, s := range seqs {
		if s.Proposer {
			return &s
		}
	}
	return nil
}

func (k Keeper) GetRollappNextProposer(ctx sdk.Context, rollappId string) (seq *types.Sequencer) {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	for _, s := range seqs {
		if s.NextProposer {
			return &s
		}
	}
	return nil
}

// expectedNextProposer returns the next proposer for a rollapp
func (k Keeper) expectedNextProposer(ctx sdk.Context, rollappId string) (seq *types.Sequencer) {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	if len(seqs) == 0 {
		return
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].Tokens.IsAllGT(seqs[j].Tokens)
	})

	// filter out proposer and nextProposer
	for _, s := range seqs {
		if s.Proposer || s.NextProposer {
			continue
		}
		seq = &s
		break
	}

	return seq
}

// StartRotation sets the proposer for a rollapp to be the next sequencer in the list
// This function will not clear the current proposer
func (k Keeper) StartRotation(ctx sdk.Context, rollappId string) {
	addr := ""
	nextProposer := k.expectedNextProposer(ctx, rollappId)
	if nextProposer != nil {
		addr = nextProposer.SequencerAddress
		nextProposer.NextProposer = true
		k.UpdateSequencer(ctx, *nextProposer, types.Bonded)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotationStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeyNextProposer, addr),
		),
	)
}

// RotateProposer completes the sequencer rotation flow.
// it will start unbonding the current proposer, and set new proposer from the bonded sequencers
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) {
	propopser := k.GetRollappProposer(ctx, rollappId)
	if propopser != nil {
		propopser.Proposer = false
		_, err := k.setSequencerToUnbonding(ctx, propopser)
		if err != nil {
			k.Logger(ctx).Error("unbond proposer", "err", err)
			// fixme: return err?
		}
	}

	addr := ""
	nextProposer := k.GetRollappNextProposer(ctx, rollappId)
	if nextProposer == nil {
		k.Logger(ctx).Error("no next proposer available", "rollappId", rollappId)
	} else {
		addr = nextProposer.SequencerAddress
		nextProposer.Proposer = true
		nextProposer.NextProposer = false
		k.UpdateSequencer(ctx, *nextProposer, types.Bonded)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, addr),
		),
	)
}
