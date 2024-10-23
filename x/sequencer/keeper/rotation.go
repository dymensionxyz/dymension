package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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
		if !k.IsSuccessor(ctx, seq) {
			// next proposer cannot mature it's notice period until the current proposer has finished rotation
			// minor effect as notice_period >>> rotation time
			k.removeNoticePeriodSequencer(ctx, seq)
			k.startRotation(ctx, seq.RollappId)

		}
	}
}

/* -------------------------------------------------------------------------- */
/*                                notice period                               */
/* -------------------------------------------------------------------------- */

// GetMatureNoticePeriodSequencers returns all sequencers that have finished their notice period
func (k Keeper) GetMatureNoticePeriodSequencers(ctx sdk.Context, endTime time.Time) (list []types.Sequencer) {
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

// AddSequencerToNoticePeriodQueue set sequencer in notice period queue
func (k Keeper) AddSequencerToNoticePeriodQueue(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)

	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.Address, sequencer.NoticePeriodTime)
	store.Set(noticePeriodKey, b)
}

// remove sequencer from notice period queue
func (k Keeper) removeNoticePeriodSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.Address, sequencer.NoticePeriodTime)
	store.Delete(noticePeriodKey)
}
