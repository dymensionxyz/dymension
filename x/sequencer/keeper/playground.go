package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// ~~~~~~~~~~~~~~~
// PLAYGROUND ONLY
// ~~~~~~~~~~~~~~~

func (k Keeper) Prune(ctx sdk.Context) {

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.SequencerKey(types.SentinelSeqAddr))

	for _, status := range types.AllStatus {
		oldKey := types.SequencerByRollappByStatusKey("", types.SentinelSeqAddr, status)
		ctx.KVStore(k.storeKey).Delete(oldKey)
	}

}
