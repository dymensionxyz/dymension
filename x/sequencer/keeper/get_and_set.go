package keeper

import (
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// SetSequencer : write to store indexed by address, and also by status
// Note: do not call with sentinel sequencer
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

// SetSequencerByDymintAddr : allows reverse lookup of sequencer by dymint address
func (k Keeper) SetSequencerByDymintAddr(ctx sdk.Context, dymint cryptotypes.Address, addr string) error {
	// could move this inside SetSequencer but it would require propogating error up a lot
	return k.dymintProposerAddrToAccAddr.Set(ctx, dymint, addr)
}

// SetProposer : passing sentinel is allowed
func (k Keeper) SetProposer(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	activeKey := types.ProposerByRollappKey(rollapp)
	store.Set(activeKey, addressBytes)
}

// SetSuccessor : passing sentinel is allowed
func (k Keeper) SetSuccessor(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.SuccessorByRollappKey(rollapp)
	store.Set(nextProposerKey, addressBytes)
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

func (k Keeper) RollappSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappKey(rollappId))
}

func (k Keeper) RollappSequencersPaginated(ctx sdk.Context, rollappId string, pageReq *query.PageRequest) ([]types.Sequencer, *query.PageResponse, error) {
	return k.prefixSequencersPaginated(ctx, types.SequencersByRollappKey(rollappId), pageReq)
}

func (k Keeper) RollappSequencersByStatus(ctx sdk.Context, rollappId string, status types.OperatingStatus) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersByRollappByStatusKey(rollappId, status))
}

func (k Keeper) RollappSequencersByStatusPaginated(ctx sdk.Context, rollappId string, status types.OperatingStatus, pageReq *query.PageRequest) ([]types.Sequencer, *query.PageResponse, error) {
	return k.prefixSequencersPaginated(ctx, types.SequencersByRollappByStatusKey(rollappId, status), pageReq)
}

func (k Keeper) RollappBondedSequencers(ctx sdk.Context, rollappId string) []types.Sequencer {
	return k.RollappSequencersByStatus(ctx, rollappId, types.Bonded)
}

func (k Keeper) AllSequencers(ctx sdk.Context) []types.Sequencer {
	return k.prefixSequencers(ctx, types.SequencersKeyPrefix)
}

func (k Keeper) prefixSequencers(ctx sdk.Context, prefixKey []byte) []types.Sequencer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	it := sdk.KVStorePrefixIterator(store, []byte{})

	defer it.Close() // nolint: errcheck

	var ret []types.Sequencer
	for ; it.Valid(); it.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(it.Value(), &val)
		ret = append(ret, val)
	}

	return ret
}

func (k Keeper) prefixSequencersPaginated(ctx sdk.Context, prefixKey []byte, pageReq *query.PageRequest) ([]types.Sequencer, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)

	var sequencers []types.Sequencer

	pageRes, err := query.Paginate(store, pageReq, func(key []byte, value []byte) error {
		var val types.Sequencer
		if err := k.cdc.Unmarshal(value, &val); err != nil {
			return err
		}
		sequencers = append(sequencers, val)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return sequencers, pageRes, nil
}

// GetSequencer returns the sentinel sequencer if not found. Use GetRealSequencer if expecting
// to get a real sequencer.
func (k Keeper) GetSequencer(ctx sdk.Context, addr string) types.Sequencer {
	seq, err := k.RealSequencer(ctx, addr)
	if err != nil {
		return k.SentinelSequencer(ctx)
	}
	return seq
}

// RealSequencer tries to get a real (non sentinel) sequencer.
func (k Keeper) RealSequencer(ctx sdk.Context, addr string) (types.Sequencer, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		return types.Sequencer{}, types.ErrSequencerNotFound
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret, nil
}

func (k Keeper) SequencerByDymintAddr(ctx sdk.Context, addr cryptotypes.Address) (types.Sequencer, error) {
	accAddr, err := k.dymintProposerAddrToAccAddr.Get(ctx, addr)
	if err != nil {
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return types.Sequencer{}, gerrc.ErrNotFound
		}
		return types.Sequencer{}, err
	}
	return k.RealSequencer(ctx, accAddr)
}

func (k Keeper) AllProposers(ctx sdk.Context) (list []types.Sequencer) {
	return k.prefixSequencerAddrs(ctx, types.ProposerByRollappKey(""))
}

func (k Keeper) AllSuccessors(ctx sdk.Context) []types.Sequencer {
	return k.prefixSequencerAddrs(ctx, types.SuccessorByRollappKey(""))
}

func (k Keeper) prefixSequencerAddrs(ctx sdk.Context, pref []byte) []types.Sequencer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), pref)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck
	ret := []types.Sequencer{}
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Value())
		seq := k.GetSequencer(ctx, address)
		ret = append(ret, seq)
	}
	return ret
}

func (k Keeper) GetProposer(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposerByRollappKey(rollapp))
	if bz == nil {
		return k.SentinelSequencer(ctx)
	}
	return k.GetSequencer(ctx, string(bz))
}

func (k Keeper) GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SuccessorByRollappKey(rollapp))
	if bz == nil {
		return k.SentinelSequencer(ctx)
	}
	return k.GetSequencer(ctx, string(bz))
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
		seq, err := k.RealSequencer(ctx, string(iterator.Value()))
		if err != nil {
			return nil, gerrc.ErrInternal.Wrapf("sequencer in notice queue but missing sequencer object: addr: %s", addr)
		}
		ret = append(ret, seq)
	}

	return ret, nil
}
