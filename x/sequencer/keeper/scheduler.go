package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// SetScheduler set a specific scheduler in the store from its index
func (k Keeper) SetScheduler(ctx sdk.Context, scheduler types.Scheduler) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SchedulerKeyPrefix))
	b := k.cdc.MustMarshal(&scheduler)
	store.Set(types.SchedulerKey(
		scheduler.SequencerAddress,
	), b)
}

// GetScheduler returns a scheduler from its index
func (k Keeper) GetScheduler(
	ctx sdk.Context,
	sequencerAddress string,

) (val types.Scheduler, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SchedulerKeyPrefix))

	b := store.Get(types.SchedulerKey(
		sequencerAddress,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveScheduler removes a scheduler from the store
func (k Keeper) RemoveScheduler(
	ctx sdk.Context,
	sequencerAddress string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SchedulerKeyPrefix))
	store.Delete(types.SchedulerKey(
		sequencerAddress,
	))
}

// GetAllScheduler returns all scheduler
func (k Keeper) GetAllScheduler(ctx sdk.Context) (list []types.Scheduler) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SchedulerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Scheduler
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
