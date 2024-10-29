package keeper

import (
	"slices"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) RollappSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappKey(rollappId))
}

func (k Keeper) RollappSequencersByStatus(ctx sdk.Context, rollappId string, status types.OperatingStatus) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappByStatusKey(rollappId, status))
}

func (k Keeper) RollappBondedSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.RollappSequencersByStatus(ctx, rollappId, types.Bonded)
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

func (k Keeper) RollappPotentialProposers(ctx sdk.Context, rollappId string) []types.Sequencer {
	seqs := k.RollappBondedSequencers(ctx, rollappId)
	seqs = slices.DeleteFunc(seqs, func(seq types.Sequencer) bool {
		return !k.isPotentialProposer(ctx, seq)
	})
	return append(seqs, k.SentinelSequencer(ctx))
}

func (k Keeper) AllSequencers(ctx sdk.Context) (list []types.Sequencer) {
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
	s, _ := k.GetRealSequencer(ctx, addr)
	return s
}

func (k Keeper) GetSequencer(ctx sdk.Context, addr string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		// TODO: possible case?
		return k.SentinelSequencer(ctx)
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret
}

// TODO: could change to OK api
func (k Keeper) GetRealSequencer(ctx sdk.Context, addr string) (types.Sequencer, error) {
	s := k.GetSequencer(ctx, addr)
	if s.Sentinel() {
		return types.Sequencer{}, types.ErrSequencerNotFound
	}
	return s, nil
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

func (k Keeper) SequencerByDymintAddr(ctx sdk.Context, addr cryptotypes.Address) (types.Sequencer, error) {
	accAddr, err := k.dymintProposerAddrToAccAddr.Get(ctx, addr)
	if err != nil {
		return types.Sequencer{}, err
	}
	return k.GetRealSequencer(ctx, accAddr)
}

func (k Keeper) SetSequencerByDymintAddr(ctx sdk.Context, dymint cryptotypes.Address, hub string) error {
	return k.dymintProposerAddrToAccAddr.Set(ctx, dymint, hub)
}

// AllProposers returns all proposers for all rollapps
// TODO: check sentinel
func (k Keeper) AllProposers(ctx sdk.Context) (list []types.Sequencer) {
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
		return k.GetSequencer(ctx, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, string(bz))
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
		return k.GetSequencer(ctx, types.SentinelSeqAddr)
	}
	return k.GetSequencer(ctx, string(bz))
}

func (k Keeper) SetSuccessor(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.SuccessorByRollappKey(rollapp)
	store.Set(nextProposerKey, addressBytes)
}
