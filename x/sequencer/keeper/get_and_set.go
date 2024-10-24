package keeper

import (
	"slices"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

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

func (k Keeper) GetRollappPotentialProposers(ctx sdk.Context, rollappId string) []types.Sequencer {
	seqs := k.GetRollappBondedSequencers(ctx, rollappId)
	slices.DeleteFunc(seqs, func(s types.Sequencer) bool {
		return !s.OptedIn
	})
	return seqs
}

func (k Keeper) GetRollappBondedSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
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

func (k Keeper) MustGetNonSentinelSequencer(ctx sdk.Context, addr string) types.Sequencer {
	s, _ := k.tryGetSequencer(ctx, addr)
	return s
}

func (k Keeper) GetSequencer(ctx sdk.Context, rollapp, addr string) types.Sequencer {
	if addr == types.SentinelSeqAddr {
		return types.SentinelSequencer(rollapp)
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		// TODO: possible case?
		return k.GetSequencer(ctx, rollapp, types.SentinelSeqAddr)
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret
}

func (k Keeper) tryGetSequencer(ctx sdk.Context, addr string) (types.Sequencer, error) {
	if addr == types.SentinelSeqAddr {
		return types.Sequencer{}, gerrc.ErrInternal.Wrap("try get sequencer only to be used on external arguments")
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		return types.Sequencer{}, types.ErrSequencerNotFound
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret, nil
}

// SetSequencer set a specific sequencer in the store from its index
func (k Keeper) SetSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(
		sequencer.Address,
	), b)

	seqByRollappKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, sequencer.Status)
	store.Set(seqByRollappKey, b)
}

// UpdateSequencerLeg updates the state of a sequencer in the keeper.
// Parameters:
//   - sequencer: The sequencer object to be updated.
//   - oldStatus: An optional parameter representing the old status of the sequencer.
//     Needs to be provided if the status of the sequencer has changed (e.g from Bonded to Unbonding).
func (k Keeper) UpdateSequencerLeg(ctx sdk.Context, sequencer *types.Sequencer, oldStatus *types.OperatingStatus) {
	k.SetSequencer(ctx, *sequencer)

	// status changed, need to remove old status key
	if oldStatus != nil {
		oldKey := types.SequencerByRollappByStatusKey(sequencer.RollappId, sequencer.Address, *oldStatus)
		ctx.KVStore(k.storeKey).Delete(oldKey)
	}
}

// GetAllProposers returns all proposers for all rollapps
// TODO: doesn't include sentinel
func (k Keeper) GetAllProposers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposerByRollappKey(""))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq := k.MustGetNonSentinelSequencer(ctx, address)
		list = append(list, seq)
	}

	return
}

func (k Keeper) GetProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposerByRollappKey(rollappId))
	if bz == nil {
		return k.GetSequencer(ctx, rollappId, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, rollappId, string(bz))
}

func (k Keeper) SetProposer(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)

	activeKey := types.ProposerByRollappKey(rollapp)
	store.Set(activeKey, addressBytes)
}

func (k Keeper) GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextProposerByRollappKey(rollapp))
	if bz == nil {
		return k.GetSequencer(ctx, rollapp, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, rollapp, string(bz))
}

func (k Keeper) SetSuccessor(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.NextProposerByRollappKey(rollapp)
	store.Set(nextProposerKey, addressBytes)
}
