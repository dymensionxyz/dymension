package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// GetMatureDecreasingBondSequencers returns all decreasing bond items for the given time
func (k Keeper) GetMatureDecreasingBondSequencers(ctx sdk.Context, endTime time.Time) (unbondings []types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.DecreasingBondQueueKey, sdk.PrefixEndBytes(types.DecreasingBondQueueByTimeKey(endTime)))
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		var b types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &b)
		unbondings = append(unbondings, b)
	}
	return
}

// SetDecreasingBondQueue sets the bond reduction item in the decreasing bond queue
func (k Keeper) SetDecreasingBondQueue(ctx sdk.Context, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&bondReduction)

	unbondingQueueKey := types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.GetUnbondTime())
	store.Set(unbondingQueueKey, b)
}

// removeDecreasingBondQueue removes the bond reduction item from the decreasing bond queue
func (k Keeper) removeDecreasingBondQueue(ctx sdk.Context, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	unbondingQueueKey := types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.GetUnbondTime())
	store.Delete(unbondingQueueKey)
}

// getSequencerDecreasingBonds returns the bond reduction item given sequencer address
func (k Keeper) getSequencerDecreasingBonds(ctx sdk.Context, sequencerAddr string) (bds []types.BondReduction) {
	prefixKey := types.DecreasingBondQueueKey
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var bd types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &bd)
		if bd.SequencerAddress == sequencerAddr {
			bds = append(bds, bd)
		}
	}

	return
}

func (k Keeper) GetAllBondReductions(ctx sdk.Context) (bds []types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DecreasingBondQueueKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		var bd types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &bd)
		bds = append(bds, bd)
	}
	return

}
