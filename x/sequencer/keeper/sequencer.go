package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// SetSequencer set a specific sequencer in the store from its index
func (k Keeper) SetSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(
		sequencer.SequencerAddress,
	), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.SequencerAddress, sequencer.Status)
	store.Set(seqByRollappKey, b)
}

func (k Keeper) UpdateSequencer(ctx sdk.Context, sequencer types.Sequencer, oldStatus types.OperatingStatus) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(sequencer.SequencerAddress), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.SequencerAddress, sequencer.Status)
	store.Set(seqByRollappKey, b)

	// status changed, need to remove old status key
	if sequencer.Status != oldStatus {
		oldKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.SequencerAddress, oldStatus)
		store.Delete(oldKey)
	}
}

// GetSequencer returns a sequencer from its index
func (k Keeper) GetSequencer(ctx sdk.Context, sequencerAddress string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(
		sequencerAddress,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllSequencers returns all sequencer
func (k Keeper) GetAllSequencers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SequencersKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSequencersByRollapp returns a sequencersByRollapp from its index
func (k Keeper) GetSequencersByRollapp(ctx sdk.Context, rollappId string) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SequencersByRollappKey(rollappId))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSequencersByRollappByStatus returns a sequencersByRollapp from its index
func (k Keeper) GetSequencersByRollappByStatus(ctx sdk.Context, rollappId string, status types.OperatingStatus) (list []types.Sequencer) {
	prefixKey := types.SequencersByRollappByStatusKey(rollappId, status)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

/* -------------------------------------------------------------------------- */
/*                               Unbonding queue                              */
/* -------------------------------------------------------------------------- */
// GetMatureUnbondingSequencers returns all unbonding sequencers
func (k Keeper) GetMatureUnbondingSequencers(ctx sdk.Context, endTime time.Time) (list []types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.UnbondingQueueKey, sdk.PrefixEndBytes(types.UnbondingQueueByTimeKey(endTime)))

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) SetUnbondingSequencerQueue(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)

	unbondingQueueKey := types.UnbondingSequencerKey(sequencer.SequencerAddress, sequencer.UnbondTime)
	store.Set(unbondingQueueKey, b)
}

// remove unbonding sequencer from the queue
func (k Keeper) removeUnbondingSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	unbondingQueueKey := types.UnbondingSequencerKey(sequencer.SequencerAddress, sequencer.UnbondTime)
	store.Delete(unbondingQueueKey)
}

/* -------------------------------------------------------------------------- */
/*                                notice period                               */
/* -------------------------------------------------------------------------- */
// get finished notice period sequencers
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

// set sequencer to notice period
func (k Keeper) SetNoticePeriodQueue(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)

	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.SequencerAddress, sequencer.UnbondTime)
	store.Set(noticePeriodKey, b)
}
