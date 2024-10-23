package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) SetRegisteredDenom(ctx sdk.Context, rollappID, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyRegisteredDenomPrefix))
	key := types.KeyRegisteredDenom(rollappID, denom)
	store.Set(key, []byte{})
}

func (k Keeper) HasRegisteredDenom(ctx sdk.Context, rollappID, denom string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyRegisteredDenomPrefix))
	key := types.KeyRegisteredDenom(rollappID, denom)
	return store.Has(key)
}

func (k Keeper) GetAllRegisteredDenoms(ctx sdk.Context, rollappID string) []string {
	var denoms []string
	k.IterateRegisteredDenoms(ctx, rollappID, func(denom string) bool {
		denoms = append(denoms, denom)
		return false
	})
	return denoms
}

func (k Keeper) IterateRegisteredDenoms(ctx sdk.Context, rollappID string, cb func(denom string) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyRegisteredDenomPrefix))
	pref := types.RegisteredDenomPrefix(rollappID)
	iterator := sdk.KVStorePrefixIterator(store, pref)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		denom := string(key[len(pref):])
		if cb(denom) {
			break
		}
	}
}
