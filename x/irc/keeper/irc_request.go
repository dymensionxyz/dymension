package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/irc/types"
)

// SetIRCRequest set a specific ircRequest in the store from its index
func (k Keeper) SetIRCRequest(ctx sdk.Context, ircRequest types.IRCRequest) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IRCRequestKeyPrefix))
	b := k.cdc.MustMarshal(&ircRequest)
	store.Set(types.IRCRequestKey(
		ircRequest.ReqId,
	), b)
}

// GetIRCRequest returns a ircRequest from its index
func (k Keeper) GetIRCRequest(
	ctx sdk.Context,
	reqId uint64,

) (val types.IRCRequest, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IRCRequestKeyPrefix))

	b := store.Get(types.IRCRequestKey(
		reqId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveIRCRequest removes a ircRequest from the store
func (k Keeper) RemoveIRCRequest(
	ctx sdk.Context,
	reqId uint64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IRCRequestKeyPrefix))
	store.Delete(types.IRCRequestKey(
		reqId,
	))
}

// GetAllIRCRequest returns all ircRequest
func (k Keeper) GetAllIRCRequest(ctx sdk.Context) (list []types.IRCRequest) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.IRCRequestKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.IRCRequest
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
