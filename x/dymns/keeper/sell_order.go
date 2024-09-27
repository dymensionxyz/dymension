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

// DEPRECATED: SetActiveSellOrdersExpiration stores the expiration of the active Sell-Orders records into the KVStore.
// When a Sell-Order is created, it has an expiration time, later be processed by the batch processing in `x/epochs` hooks,
// instead of iterating through all the Sell-Orders records in store, we store the expiration of the active Sell-Orders,
// so that we can easily find the expired Sell-Orders and conditionally load them to process.
// This expiration list should be maintained accordingly to the Sell-Order CRUD.
func (k Keeper) SetActiveSellOrdersExpiration(ctx sdk.Context,
	activeSellOrdersExpiration *dymnstypes.ActiveSellOrdersExpiration, assetType dymnstypes.AssetType,
) error {
	activeSellOrdersExpiration.Sort()

	if err := activeSellOrdersExpiration.Validate(); err != nil {
		return err
	}

	// use key according to asset type
	var key []byte
	switch assetType {
	case dymnstypes.TypeName:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.TypeAlias:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid asset type: " + assetType.PrettyName())
	}

	// persist record
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(activeSellOrdersExpiration)
	store.Set(key, bz)
	return nil
}

// DEPRECATED: GetActiveSellOrdersExpiration retrieves the expiration of the active Sell-Orders records
// of the corresponding asset type.
// For more information, see SetActiveSellOrdersExpiration.
func (k Keeper) GetActiveSellOrdersExpiration(ctx sdk.Context,
	assetType dymnstypes.AssetType,
) *dymnstypes.ActiveSellOrdersExpiration {
	store := ctx.KVStore(k.storeKey)

	// use key according to asset type
	var key []byte
	switch assetType {
	case dymnstypes.TypeName:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.TypeAlias:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid asset type: " + assetType.PrettyName())
	}

	var activeSellOrdersExpiration dymnstypes.ActiveSellOrdersExpiration

	bz := store.Get(key)
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &activeSellOrdersExpiration)
	}

	if activeSellOrdersExpiration.Records == nil {
		activeSellOrdersExpiration.Records = make([]dymnstypes.ActiveSellOrdersExpirationRecord, 0)
	}

	return &activeSellOrdersExpiration
}
