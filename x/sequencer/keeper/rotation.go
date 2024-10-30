package keeper

import (
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// StartNoticePeriodForSequencer defines a period of time for the sequencer where
// they cannot yet unbond, nor submit their last block. Adds to a queue for later
// processing.
func (k Keeper) StartNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) {
	seq.NoticePeriodTime = ctx.BlockTime().Add(k.GetParams(ctx).NoticePeriod)

	k.AddToNoticeQueue(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, seq.NoticePeriodTime.String()),
		),
	)
}

func (k Keeper) AddToNoticeQueue(ctx sdk.Context, seq types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticeQueueBySeqTimeKey(seq.Address, seq.NoticePeriodTime)
	store.Set(noticePeriodKey, []byte(seq.Address))
}

func (k Keeper) removeFromNoticeQueue(ctx sdk.Context, seq types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticeQueueBySeqTimeKey(seq.Address, seq.NoticePeriodTime)
	store.Delete(noticePeriodKey)
}

// NoticeElapsedSequencers gets all sequencers across all rollapps whose notice period
// has passed/elapsed.
func (k Keeper) NoticeElapsedSequencers(ctx sdk.Context, endTime time.Time) ([]types.Sequencer, error) {
	ret := []types.Sequencer{}
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.NoticePeriodQueueKey, sdk.PrefixEndBytes(types.NoticeQueueByTimeKey(endTime)))

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		addr := string(iterator.Value())
		seq, err := k.GetRealSequencer(ctx, string(iterator.Value()))
		if err != nil {
			return nil, gerrc.ErrInternal.Wrapf("sequencer in notice queue but missing sequencer object: addr: %s", addr)
		}
		ret = append(ret, seq)
	}

	return ret, nil
}

// ChooseSuccessorForFinishedNotices goes through all sequencers whose notice periods have elapsed.
// For each proposer, it chooses a successor proposer for their rollapp.
func (k Keeper) ChooseSuccessorForFinishedNotices(ctx sdk.Context, now time.Time) error {
	seqs, err := k.NoticeElapsedSequencers(ctx, now)
	if err != nil {
		return errorsmod.Wrap(err, "get notice elapsed sequencers")
	}
	for _, seq := range seqs {
		// Successor cannot finish notice. The proposer must finish first and then rotate to the successor.
		if !k.IsSuccessor(ctx, seq) {
			k.removeFromNoticeQueue(ctx, seq)
			k.chooseSuccessor(ctx, seq.RollappId)
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
	}
	return nil
}

func (k Keeper) awaitingLastProposerBlock(ctx sdk.Context, rollapp string) bool {
	proposer := k.GetProposer(ctx, rollapp)
	return proposer.NoticeElapsed(ctx.BlockTime())
}

// OnProposerLastBlock : it will assign the successor to be the proposer.
func (k Keeper) OnProposerLastBlock(ctx sdk.Context, proposer types.Sequencer) error {
	allowLastBlock := proposer.NoticeElapsed(ctx.BlockTime())
	if !allowLastBlock {
		return errorsmod.Wrap(gerrc.ErrFault, "sequencer has submitted last block without finishing notice period")
	}

	k.SetProposer(ctx, proposer.RollappId, types.SentinelSeqAddr)
	if err := k.ChooseProposer(ctx, proposer.RollappId); err != nil {
		return errorsmod.Wrap(err, "choose proposer")
	}
	after := k.GetProposer(ctx, proposer.RollappId)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, proposer.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, after.Address),
		),
	)
	return nil
}
