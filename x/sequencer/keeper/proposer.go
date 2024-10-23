package keeper

import (
	"slices"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) ChooseProposer(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded - invariant broken")
		}
	}
	successor := k.GetSuccessor(ctx, rollapp)
	k.SetProposer(ctx, rollapp, successor.Address)
	k.SetSuccessor(ctx, rollapp, types.SentinelSequencerAddr)
	if k.GetProposer(ctx, rollapp).Sentinel() {
		seqs := k.GetRollappPotentialProposers(ctx, rollapp)
		slices.DeleteFunc(seqs, func(s types.Sequencer) bool { // Not efficient, could optimize.
			return s.Address == proposer.Address
		})
		// TODO: exclude last? thats what the legacy code does
		proposer := proposerChoiceAlgo(rollapp, seqs)
		k.SetProposer(ctx, rollapp, proposer.Address)
	}
	return nil
}

func (k Keeper) ChooseSuccessor(ctx sdk.Context, rollapp string) error {
	proposer := k.GetProposer(ctx, rollapp)
	if proposer.Sentinel() {
		return gerrc.ErrInternal.Wrap("can not choose successor if proposer is sentinel")
	}
	successor := k.GetSuccessor(ctx, rollapp)
	if successor.Sentinel() {
		seqs := k.GetRollappPotentialProposers(ctx, rollapp)
		slices.DeleteFunc(seqs, func(s types.Sequencer) bool { // Not efficient, could optimize.
			return s.Address == proposer.Address
		})
		successor := proposerChoiceAlgo(rollapp, seqs)
		k.SetSuccessor(ctx, rollapp, successor.Address)

	}
	return nil
}

func proposerChoiceAlgo(rollapp string, seqs []types.Sequencer) types.Sequencer {
	if len(seqs) == 0 {
		return types.SentinelSequencer(rollapp)
	}
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].TokensCoin().IsGTE(seqs[j].TokensCoin())
	})
	return seqs[0]
}

func (k Keeper) GetProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposerByRollappKey(rollappId))
	if bz == nil {
		return k.GetSequencer(ctx, rollappId, types.SentinelSequencerAddr)
	}
	return k.GetSequencer(ctx, rollappId, string(bz))
}

func (k Keeper) GetSuccessor(ctx sdk.Context, rollapp string) types.Sequencer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextProposerByRollappKey(rollapp))
	if bz == nil {
		return k.GetSequencer(ctx, rollapp, types.SentinelSequencerAddr)
	}
	return k.GetSequencer(ctx, rollapp, string(bz))
}

func (k Keeper) SetProposer(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)

	activeKey := types.ProposerByRollappKey(rollapp)
	store.Set(activeKey, addressBytes)
}

func (k Keeper) SetSuccessor(ctx sdk.Context, rollapp, seqAddr string) {
	store := ctx.KVStore(k.storeKey)
	addressBytes := []byte(seqAddr)
	nextProposerKey := types.NextProposerByRollappKey(rollapp)
	store.Set(nextProposerKey, addressBytes)
}

func (k Keeper) IsSuccessor(ctx sdk.Context, seq types.Sequencer) bool {
	return seq.Address == k.GetSuccessor(ctx, seq.RollappId).Address
}

func (k Keeper) TryGetSequencer(ctx sdk.Context, addr string) (types.Sequencer, error) {
	if addr == types.SentinelSequencerAddr {
		return types.Sequencer{}, gerrc.ErrInternal.Wrap("try get sequencer only to be used on external arguments")
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		return types.Sequencer{}, types.ErrSequencerNotFound
	}
	// rollapp arg not needed since it's only needed to create sentinel seq, which we definitely won't do
	return k.GetSequencer(ctx, "", addr), nil
}

func (k Keeper) GetSequencer(ctx sdk.Context, rollapp, addr string) types.Sequencer {
	if addr == types.SentinelSequencerAddr {
		return types.SentinelSequencer(rollapp)
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.SequencerKey(addr))
	if b == nil {
		// TODO: possible case?
		return k.GetSequencer(ctx, rollapp, types.SentinelSequencerAddr)
	}
	ret := types.Sequencer{}
	k.cdc.MustUnmarshal(b, &ret)
	return ret
}
