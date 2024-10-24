package keeper

import (
	"slices"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) RollappSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappKey(rollappId))
}

func (k Keeper) RollappSequencersByStatus(ctx sdk.Context, rollappId string, status types.OperatingStatus) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappByStatusKey(rollappId, status))
}

func (k Keeper) prefixSequencers(ctx sdk.Context, prefixKey []byte) []types.Sequencer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	it := sdk.KVStorePrefixIterator(store, []byte{})

	defer it.Close() // nolint: errcheck

	ret := []types.Sequencer{}
	for ; it.Valid(); it.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(it.Value(), &val)
		ret = append(ret, val)
	}

	return ret
}

func (k Keeper) GetRollappPotentialProposers(ctx sdk.Context, rollappId string) []types.Sequencer {
	seqs := k.GetRollappBondedSequencers(ctx, rollappId)
	slices.DeleteFunc(seqs, func(s types.Sequencer) bool {
		return !s.OptedIn
	})
	return seqs
}

func (k Keeper) GetRollappBondedSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.RollappSequencersByStatus(ctx, rollappId, types.Bonded)
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

func (k Keeper) SetSequencer(ctx sdk.Context, seq types.Sequencer) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&seq)
	store.Set(types.SequencerKey(seq.Address), b)

	for _, status := range types.AllStatus {
		oldKey := types.SequencerByRollappByStatusKey(seq.RollappId, seq.Address, status)
		ctx.KVStore(k.storeKey).Delete(oldKey)
	}

	seqByRollappKey := types.SequencerByRollappByStatusKey(seq.RollappId, seq.Address, seq.Status)
	store.Set(seqByRollappKey, b)
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

func (k Keeper) GetProposer(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposerByRollappKey(rollapp))
	if bz == nil {
		return k.GetSequencer(ctx, rollapp, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, rollapp, string(bz))
}

func (k Keeper) SetProposer(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)

	activeKey := types.ProposerByRollappKey(rollapp)
	store.Set(activeKey, addressBytes)
}

func (k Keeper) GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SuccessorByRollappKey(rollapp))
	if bz == nil {
		return k.GetSequencer(ctx, rollapp, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, rollapp, string(bz))
}

func (k Keeper) SetSuccessor(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.SuccessorByRollappKey(rollapp)
	store.Set(nextProposerKey, addressBytes)
}
