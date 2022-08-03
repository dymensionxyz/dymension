package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
)

// SetSequencer set a specific sequencer in the store from its index
func (k Keeper) SetSequencer(ctx sdk.Context, sequencer types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))
	b := k.cdc.MustMarshal(&sequencer)
	store.Set(types.SequencerKey(
		sequencer.SequencerAddress,
	), b)
}

// GetSequencer returns a sequencer from its index
func (k Keeper) GetSequencer(
	ctx sdk.Context,
	sequencerAddress string,

) (val types.Sequencer, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))

	b := store.Get(types.SequencerKey(
		sequencerAddress,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSequencer removes a sequencer from the store
func (k Keeper) RemoveSequencer(
	ctx sdk.Context,
	sequencerAddress string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))
	store.Delete(types.SequencerKey(
		sequencerAddress,
	))
}

// GetAllSequencer returns all sequencer
func (k Keeper) GetAllSequencer(ctx sdk.Context) (list []types.Sequencer) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	// nolint: errcheck
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Sequencer
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
