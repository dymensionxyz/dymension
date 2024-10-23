package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) ChooseProposer(ctx sdk.Context, rollappId string) error {
	proposer, err := k.GetProposer(ctx, rollappId)
	if err != nil {
		return errorsmod.Wrap(err, "get proposer")
	}
	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded")
		}
	}
	successor := k.GetProposer(ctx, rollappId)
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
