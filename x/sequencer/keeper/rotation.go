package keeper

import (
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
		if err := k.chooseSuccessor(ctx, seq.RollappId); err != nil {
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
	k.SetProposer(ctx, rollapp, successor.Address)
	k.SetSuccessor(ctx, rollapp, types.SentinelSeqAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, proposer.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, successor.Address),
		),
	)
	return nil
}
