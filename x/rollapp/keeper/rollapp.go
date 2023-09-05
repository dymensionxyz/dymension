package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// SetRollapp set a specific rollapp in the store from its index
func (k Keeper) SetRollapp(ctx sdk.Context, rollapp types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	b := k.cdc.MustMarshal(&rollapp)
	store.Set(types.RollappKey(
		rollapp.RollappId,
	), b)

	// check if chain-id is EVM compatible
	eip155, err := types.ParseChainID(rollapp.RollappId)
	if err != nil || eip155 == nil {
		return
	}

	store = prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))
	b = k.cdc.MustMarshal(&rollapp)
	store.Set(types.RollappByEIP155Key(
		eip155.Uint64(),
	), b)
}

// GetRollappByEIP155 returns a rollapp from its index
func (k Keeper) GetRollappByEIP155(
	ctx sdk.Context,
	eip155 uint64,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))

	b := store.Get(types.RollappByEIP155Key(
		eip155,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetRollapp returns a rollapp from its index
func (k Keeper) GetRollapp(
	ctx sdk.Context,
	rollappId string,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))

	b := store.Get(types.RollappKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveRollapp removes a rollapp from the store
func (k Keeper) RemoveRollapp(
	ctx sdk.Context,
	rollappId string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	store.Delete(types.RollappKey(
		rollappId,
	))
}

// GetAllRollapp returns all rollapp
func (k Keeper) GetAllRollapp(ctx sdk.Context) (list []types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	// nolint: errcheck
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

/* ------------------------- Rollapp by IBC channel ------------------------- */
// SetChannelID sets the channelID for a rollapp
func (k Keeper) SetRollappByIBCChannel(ctx sdk.Context, rollappID, portID, channelID string) {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByIBCChannelKeyPrefix))
	store.Set(types.RollappByIBCChannelKey(
		portID,
		channelID,
	), k.cdc.MustMarshal(&rollapp))
}

func (k Keeper) GetRollappByIBCChannel(
	ctx sdk.Context,
	portID string,
	channelID string,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByIBCChannelKeyPrefix))

	b := store.Get(types.RollappByIBCChannelKey(
		portID,
		channelID,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
