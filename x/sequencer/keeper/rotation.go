package keeper

import (
	"slices"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// StartNoticePeriod defines a period of time for the proposer where
// they cannot yet unbond, nor submit their last block. Adds to a queue for later
// processing.
func (k Keeper) StartNoticePeriod(ctx sdk.Context, prop *types.Sequencer) {
	prop.NoticePeriodTime = ctx.BlockTime().Add(k.GetParams(ctx).NoticePeriod)

	k.AddToNoticeQueue(ctx, *prop)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, prop.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, prop.Address),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, prop.NoticePeriodTime.String()),
		),
	)
}

// NoticeElapsedProposers gets all sequencers across all rollapps whose notice period
// has passed/elapsed.
func (k Keeper) NoticeElapsedProposers(ctx sdk.Context, endTime time.Time) ([]types.Sequencer, error) {
	return k.NoticeQueue(ctx, &endTime)
}

// ChooseSuccessorForFinishedNotices goes through all sequencers whose notice periods have elapsed.
// For each proposer, it chooses a successor proposer for their rollapp.
// Contract: must be called before OnProposerLastBlock for a given block time
func (k Keeper) ChooseSuccessorForFinishedNotices(ctx sdk.Context, now time.Time) error {
	seqs, err := k.NoticeElapsedProposers(ctx, now)
	if err != nil {
		return errorsmod.Wrap(err, "get notice elapsed sequencers")
	}
	for _, seq := range seqs {
		k.removeFromNoticeQueue(ctx, seq)
		if err := k.setSuccessorForRotatingRollapp(ctx, seq.RollappId); err != nil {
			return errorsmod.Wrap(err, "choose successor")
		}
		successor := k.GetSuccessor(ctx, seq.RollappId)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRotationStarted,
				sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
				sdk.NewAttribute(types.AttributeKeyNextProposer, successor.Address),
				sdk.NewAttribute(types.AttributeKeyRewardAddr, successor.RewardAddr),
				sdk.NewAttribute(types.AttributeKeyWhitelistedRelayers, strings.Join(successor.WhitelistedRelayers, ",")),
			),
		)
	}
	return nil
}

func (k Keeper) RotationInProgress(ctx sdk.Context, rollapp string) bool {
	prop := k.GetProposer(ctx, rollapp)
	return prop.NoticeInProgress(ctx.BlockTime()) || k.AwaitingLastProposerBlock(ctx, rollapp)
}

func (k Keeper) AwaitingLastProposerBlock(ctx sdk.Context, rollapp string) bool {
	proposer := k.GetProposer(ctx, rollapp)
	return proposer.NoticeElapsed(ctx.BlockTime())
}

// OnProposerLastBlock : it will assign the successor to be the proposer.
// Contract: must be called after ChooseSuccessorForFinishedNotices for a given block time
func (k Keeper) OnProposerLastBlock(ctx sdk.Context, proposer types.Sequencer) error {
	allowLastBlock := proposer.NoticeElapsed(ctx.BlockTime())
	if !allowLastBlock {
		return errorsmod.Wrap(gerrc.ErrFault, "sequencer has submitted last block without finishing notice period")
	}

	rollapp := proposer.RollappId

	successor := k.GetSuccessor(ctx, rollapp)
	k.SetSuccessor(ctx, rollapp, types.SentinelSeqAddr) // clear successor
	k.SetProposer(ctx, rollapp, successor.Address)

	// if successor is sentinel, prepare new revision for the rollapp
	if successor.Sentinel() {
		err := k.rollappKeeper.HardForkToLatest(ctx, rollapp)
		if err != nil {
			return errorsmod.Wrap(err, "hard fork to latest")
		}
	} else if err := k.hooks.AfterSetRealProposer(ctx, rollapp, successor); err != nil {
		return errorsmod.Wrap(err, "after set real sequencer")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, proposer.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, successor.Address),
		),
	)
	return nil
}

// setSuccessorForRotatingRollapp will assign a successor to the rollapp.
// It will prioritize non sentinel
// called when a proposer has finished their notice period.
func (k Keeper) setSuccessorForRotatingRollapp(ctx sdk.Context, rollapp string) error {
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
