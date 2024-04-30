package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// SetSequencersByRollapp set a specific sequencersByRollapp in the store from its index
func (k Keeper) SetSequencersByRollapp(ctx sdk.Context, sequencersByRollapp types.SequencersByRollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencersByRollappKeyPrefix))
	b := k.cdc.MustMarshal(&sequencersByRollapp)
	store.Set(types.SequencersByRollappKey(
		sequencersByRollapp.RollappId,
	), b)
}

// GetSequencersByRollapp returns a sequencersByRollapp from its index
func (k Keeper) GetSequencersByRollapp(
	ctx sdk.Context,
	rollappId string,

) (val types.SequencersByRollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencersByRollappKeyPrefix))

	b := store.Get(types.SequencersByRollappKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSequencersByRollapp removes a sequencersByRollapp from the store
func (k Keeper) RemoveSequencersByRollapp(
	ctx sdk.Context,
	rollappId string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencersByRollappKeyPrefix))
	store.Delete(types.SequencersByRollappKey(
		rollappId,
	))
}

// GetAllSequencersByRollapp returns all sequencersByRollapp
func (k Keeper) GetAllSequencersByRollapp(ctx sdk.Context) (list []types.SequencersByRollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SequencersByRollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.SequencersByRollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
