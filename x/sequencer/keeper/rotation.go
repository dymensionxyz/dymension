package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) time.Time {
	seq.NoticePeriodTime = ctx.BlockTime().Add(k.NoticePeriod(ctx))

	k.AddToNoticeQueue(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, seq.NoticePeriodTime.String()),
		),
	)

	return seq.NoticePeriodTime
}

// MatureSequencersWithNoticePeriod start rotation flow for all sequencers that have finished their notice period
// The next proposer is set to the next bonded sequencer
// The hub will expect a "last state update" from the sequencer to start unbonding
// In the middle of rotation, the next proposer required a notice period as well.
func (k Keeper) MatureSequencersWithNoticePeriod(ctx sdk.Context, now time.Time) {
	seqs := k.GetNoticeElapsedSequencers(ctx, now)
	for _, seq := range seqs {
		if !k.isSuccessor(ctx, seq) {
			// next proposer cannot mature its notice period until the current proposer has finished rotation
			// minor effect as notice_period >>> rotation time
			k.removeFromNoticeQueue(ctx, seq)
			if err := k.chooseSuccessor(ctx, seq.RollappId); err != nil {
				k.Logger(ctx).Error("Choose successor.", "err", err)
				continue
			}
			successor := k.GetSuccessor(ctx, seq.RollappId)
			// TODO: event cleanup
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRotationStarted,
					sdk.NewAttribute(types.AttributeKeyRollappId, seq.RollappId),
					sdk.NewAttribute(types.AttributeKeyNextProposer, successor.Address),
				),
			)
		}
	}
}

func (k Keeper) onProposerLastBlock(ctx sdk.Context, proposer types.Sequencer) error {
	allowLastBlock := proposer.NoticeElapsed(ctx.BlockTime())
	if !allowLastBlock {
		return errorsmod.Wrap(gerrc.ErrFault, "sequencer has submitted last block without finishing notice period")
	}
	k.SetProposer(ctx, proposer.RollappId, types.SentinelSeqAddr)
	k.chooseProposer(ctx, proposer.RollappId)
	return nil
}

func (k Keeper) GetNoticeElapsedSequencers(ctx sdk.Context, endTime time.Time) (list []types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.NoticePeriodQueueKey, sdk.PrefixEndBytes(types.NoticePeriodQueueByTimeKey(endTime)))

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) AddToNoticeQueue(ctx sdk.Context, seq types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&seq)
	noticePeriodKey := types.NoticePeriodSequencerKey(seq.Address, seq.NoticePeriodTime)
	store.Set(noticePeriodKey, b)
}

func (k Keeper) removeFromNoticeQueue(ctx sdk.Context, seq types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticePeriodSequencerKey(seq.Address, seq.NoticePeriodTime)
	store.Delete(noticePeriodKey)
}
