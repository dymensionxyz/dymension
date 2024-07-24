package keeper

import (
	"sort"
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
		sequencer.Address,
	), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, sequencer.Status)
	store.Set(seqByRollappKey, b)

	// To support InitGenesis scenario
	if sequencer.Status == types.Unbonding {
		k.setUnbondingSequencerQueue(ctx, sequencer)
	}
}

func (k Keeper) UpdateSequencer(ctx sdk.Context, sequencer types.Sequencer, oldStatus types.OperatingStatus) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(sequencer.Address), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, sequencer.Status)
	store.Set(seqByRollappKey, b)

	// status changed, need to remove old status key
	if sequencer.Status != oldStatus {
		oldKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, oldStatus)
		store.Delete(oldKey)
	}
}

// RotateProposer sets the proposer for a rollapp to be the proposer with the greatest bond
// This function will not clear the current proposer (assumes no proposer is set)
func (k Keeper) RotateProposer(ctx sdk.Context, rollappId string) {
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	if len(seqs) == 0 {
		k.Logger(ctx).Info("no bonded sequencer found for rollapp", "rollappId", rollappId)
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

	seq := seqs[0]
	seq.Proposer = true
	k.UpdateSequencer(ctx, seq, types.Bonded)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRotated,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollappId),
			sdk.NewAttribute(types.AttributeKeySequencer, seq.Address),
		),
	)
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

func (k Keeper) IsSequencerBonded(ctx sdk.Context, address string) bool {
	seq, found := k.GetSequencer(ctx, address)
	return found && seq.IsBonded()
}

func (k Keeper) GetSequencerByRollappByStatus(ctx sdk.Context, rollappId, seqAddress string, status types.OperatingStatus) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerByRollappByStatusKey(
		rollappId,
		seqAddress,
		status,
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

func (k Keeper) setUnbondingSequencerQueue(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)

	unbondingQueueKey := types.UnbondingSequencerKey(sequencer.Address, sequencer.UnbondTime)
	store.Set(unbondingQueueKey, b)
}

// remove unbonding sequencer from the queue
func (k Keeper) removeUnbondingSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	unbondingQueueKey := types.UnbondingSequencerKey(sequencer.Address, sequencer.UnbondTime)
	store.Delete(unbondingQueueKey)
}
