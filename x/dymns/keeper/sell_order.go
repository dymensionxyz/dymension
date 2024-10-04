package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// SetSellOrder stores a Sell-Order into the KVStore.
func (k Keeper) SetSellOrder(ctx sdk.Context, so dymnstypes.SellOrder) error {
	if err := so.Validate(); err != nil {
		return err
	}

	// micro optimize to save space
	if !so.HasSetSellPrice() {
		so.SellPrice = nil
	}
	if so.HighestBid != nil && len(so.HighestBid.Params) == 0 {
		so.HighestBid.Params = nil
	}

	// persist record
	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(so.AssetId, so.AssetType)
	bz := k.cdc.MustMarshal(&so)
	store.Set(soKey, bz)

	ctx.EventManager().EmitEvent(so.GetSdkEvent(dymnstypes.AttributeValueSoActionNameSet))

	return nil
}

// GetSellOrder retrieves active Sell-Order of the corresponding Dym-Name/Alias from the KVStore.
// If the Sell-Order does not exist, nil is returned.
func (k Keeper) GetSellOrder(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType,
) *dymnstypes.SellOrder {
	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(assetId, assetType)

	bz := store.Get(soKey)
	if bz == nil {
		return nil
	}

	var so dymnstypes.SellOrder
	k.cdc.MustUnmarshal(bz, &so)

	return &so
}

// DeleteSellOrder deletes the Sell-Order from the KVStore.
func (k Keeper) DeleteSellOrder(ctx sdk.Context, assetId string, assetType dymnstypes.AssetType) {
	so := k.GetSellOrder(ctx, assetId, assetType)
	if so == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(assetId, assetType)
	store.Delete(soKey)

	ctx.EventManager().EmitEvent(so.GetSdkEvent(dymnstypes.AttributeValueSoActionNameDelete))
}

// GetAllSellOrders returns all active Sell-Orders from the KVStore.
// No filter is applied.
// Store iterator is expensive so this function should be used only in Genesis and for testing purpose.
func (k Keeper) GetAllSellOrders(ctx sdk.Context) (list []dymnstypes.SellOrder) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixSellOrder)
	defer func() {
		_ = iterator.Close() // nolint: errcheck
	}()

	for ; iterator.Valid(); iterator.Next() {
		var so dymnstypes.SellOrder
		k.cdc.MustUnmarshal(iterator.Value(), &so)
		list = append(list, so)
	}

	return list
}
