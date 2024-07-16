package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

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

// SetNextProposer sets the proposer for a rollapp to be the next sequencer in the list
// This function will not clear the current proposer
func (k Keeper) SetNextProposer(ctx sdk.Context, rollappId string) string {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)

	// we need at least 2 sequencers to rotate (1 proposer, 1 nextProposer)
	if len(seqs) <= 1 {
		k.Logger(ctx).Info("no next bonded sequencer available", "rollappId", rollappId)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeNoBondedSequencer,
				sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			),
		)
		return
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].Tokens.IsAllGT(seqs[j].Tokens)
	})

	// filter out proposer and nextProposer
	var seq *types.Sequencer
	for _, s := range seqs {
		if s.Proposer || s.NextProposer {
			continue
		}
		seq = &s
		break
	}

	seq.NextProposer = true
	k.UpdateSequencer(ctx, *seq, types.Bonded)

	// TODO: emit event
}

// RotateProposer sets the proposer for a rollapp to be the proposer with the greatest bond
// This function will not clear the current proposer (assumes no proposer is set)
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) {
	propopser := k.GetRollappProposer(ctx, rollappId)
	if propopser != nil {
		propopser.Proposer = false
		k.UpdateSequencer(ctx, *propopser, types.Bonded)
	}

	/*
			// set this sequencer to unbonding
		_, err := k.setSequencerToUnbonding(ctx, &seq)
		if err != nil {
			return err
		}

	*/

	nextProposer := k.GetRollappNextProposer(ctx, rollappId)
	nextProposer.Proposer = true
	nextProposer.NextProposer = false
	k.UpdateSequencer(ctx, *nextProposer, types.Bonded)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, nextProposer.SequencerAddress),
		),
	)
}
