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

// UpdateSequencer updates the state of a sequencer in the keeper.
// Parameters:
//   - ctx: The context object that provides access to the store and other information.
//   - sequencer: The sequencer object to be updated.
//   - oldStatus: An optional parameter representing the old status of the sequencer.
//     Needs to be provided if the status of the sequencer has changed (e.g from Bonded to Unbonding).
func (k Keeper) UpdateSequencer(ctx sdk.Context, sequencer types.Sequencer, oldStatus ...types.OperatingStatus) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(sequencer.SequencerAddress), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.SequencerAddress, sequencer.Status)
	store.Set(seqByRollappKey, b)

	// status changed, need to remove old status key
	if len(oldStatus) > 0 && sequencer.Status != oldStatus[0] {
		oldKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.SequencerAddress, oldStatus[0])
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

func (k Keeper) AddSequencerToUnbondingQueue(ctx sdk.Context, sequencer types.Sequencer) {
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

	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.SequencerAddress, sequencer.NoticePeriodTime)
	store.Set(noticePeriodKey, b)
}

// remove sequencer from notice period queue
func (k Keeper) removeNoticePeriodSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	noticePeriodKey := types.NoticePeriodSequencerKey(sequencer.SequencerAddress, sequencer.NoticePeriodTime)
	store.Delete(noticePeriodKey)
}

/* ------------------------- proposer/next proposer ------------------------- */
// GetAllProposers returns all proposers for all rollapps
func (k Keeper) GetAllProposers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposerByRollappKey(""))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq, _ := k.GetSequencer(ctx, address)
		list = append(list, seq)
	}

	return
}

func (k Keeper) SetProposer(ctx sdk.Context, rollappId, sequencerAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(sequencerAddr)

	activeKey := types.ProposerByRollappKey(rollappId)
	store.Set(activeKey, addressBytes)
}

// GetProposer returns the proposer for a rollapp
func (k Keeper) GetProposer(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.ProposerByRollappKey(rollappId))
	if len(b) == 0 {
		return val, false
	}

	return k.GetSequencer(ctx, string(b))
}

func (k Keeper) removeProposer(ctx sdk.Context, rollappId string) {
	k.SetProposer(ctx, rollappId, "")
}

func (k Keeper) isProposer(ctx sdk.Context, rollappId, seqAddr string) bool {
	proposer, ok := k.GetProposer(ctx, rollappId)
	return ok && proposer.SequencerAddress == seqAddr
}

func (k Keeper) setNextProposer(ctx sdk.Context, rollappId, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.NextProposerByRollappKey(rollappId)
	store.Set(nextProposerKey, addressBytes)
}

// GetNextProposer returns the next proposer for a rollapp
// It will return found=false if the next proposer is not set
// It will return found=true if the next proposer is set, even if it's empty
func (k Keeper) GetNextProposer(ctx sdk.Context, rollappId string) (val types.Sequencer, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.NextProposerByRollappKey(rollappId))
	if b == nil {
		return val, false
	}

	address := string(b)
	if address == "" {
		return val, true
	}
	return k.GetSequencer(ctx, address)
}

func (k Keeper) hasNextProposer(ctx sdk.Context, rollappId string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.NextProposerByRollappKey(rollappId))
}

func (k Keeper) isNextProposer(ctx sdk.Context, rollappId, seqAddr string) bool {
	nextProposer, ok := k.GetNextProposer(ctx, rollappId)
	return ok && nextProposer.SequencerAddress == seqAddr
}

func (k Keeper) removeNextProposer(ctx sdk.Context, rollappId string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.NextProposerByRollappKey(rollappId))
}
