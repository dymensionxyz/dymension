package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) {
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

func (k Keeper) OnProposerLastBlock(ctx sdk.Context, proposer types.Sequencer) error {
	allowLastBlock := proposer.NoticeElapsed(ctx.BlockTime())
	if !allowLastBlock {
		return errorsmod.Wrap(gerrc.ErrFault, "sequencer has submitted last block without finishing notice period")
	}
	k.SetProposer(ctx, proposer.RollappId, types.SentinelSeqAddr)
	return errorsmod.Wrap(k.ChooseProposer(ctx, proposer.RollappId), "choose proposer")
}
