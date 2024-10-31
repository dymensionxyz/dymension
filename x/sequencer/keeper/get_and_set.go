package keeper

import (
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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

func (k Keeper) RollappBondedSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.RollappSequencersByStatus(ctx, rollappId, types.Bonded)
}

func (k Keeper) AllSequencers(ctx sdk.Context) (list []types.Sequencer) {
	return k.prefixSequencers(ctx, types.SequencersKeyPrefix)
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

func (k Keeper) MustGetNonSentinelSequencer(ctx sdk.Context, addr string) types.Sequencer {
	s, _ := k.GetRealSequencer(ctx, addr)
	return s
}

// GetSequencer returns the sentinel sequencer if not found. Use GetRealSequencer if expecting
// to get a real sequencer.
func (k Keeper) GetSequencer(ctx sdk.Context, addr string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		return k.SentinelSequencer(ctx)
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret
}

// GetRealSequencer tries to get a real (non sentinel) sequencer.
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
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return types.Sequencer{}, gerrc.ErrNotFound
		}
		return types.Sequencer{}, err
	}
	return k.GetRealSequencer(ctx, accAddr)
}

func (k Keeper) SetSequencerByDymintAddr(ctx sdk.Context, dymint cryptotypes.Address, addr string) error {
	// TODO: could move this inside SetSequencer but it would require propogating error up a lot
	return k.dymintProposerAddrToAccAddr.Set(ctx, dymint, addr)
}

// AllProposers returns all proposers for all rollapps
func (k Keeper) AllProposers(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ProposerByRollappKey(""))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq := k.GetSequencer(ctx, address)
		list = append(list, seq)
	}

	return
}

func (k Keeper) AllSuccessors(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SuccessorByRollappKey(""))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq := k.GetSequencer(ctx, address)
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

// NoticeQueue - the entire notice queue
func (k Keeper) NoticeQueue(ctx sdk.Context, endTime *time.Time) ([]types.Sequencer, error) {
	ret := []types.Sequencer{}
	store := ctx.KVStore(k.storeKey)
	prefix := types.NoticePeriodQueueKey
	if endTime != nil {
		prefix = types.NoticeQueueByTimeKey(*endTime)
	}
	iterator := store.Iterator(types.NoticePeriodQueueKey, sdk.PrefixEndBytes(prefix))

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
