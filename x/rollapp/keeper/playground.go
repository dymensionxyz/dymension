package keeper

// ~~~~~~~~~~~~~~~
// PLAYGROUND ONLY
// ~~~~~~~~~~~~~~~

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) Prune(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKeyPrefix
	iterator := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close() // nolint: errcheck

	ret := []types.LivenessEvent{}
	for ; iterator.Valid(); iterator.Next() {
		// events are stored in height non-decreasing order
		e := types.LivenessEventQueueKeyToEvent(iterator.Key())

		if ctx.BlockHeight() <= e.HubHeight {
			break
		}
		ret = append(ret, e)
	}
	for _, e := range ret {
		k.DelLivenessEvents(ctx, e.HubHeight, e.RollappId)
	}
}
