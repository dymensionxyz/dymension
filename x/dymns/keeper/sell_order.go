package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// SetSellOrder stores a Sell-Order into the KVStore.
func (k Keeper) SetSellOrder(ctx sdk.Context, so dymnstypes.SellOrder) error {
	if err := so.Validate(); err != nil {
		return err
	}

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

// MoveSellOrderToHistorical moves the active Sell-Order record of the Dym-Name/Alias
// into historical, and deletes the original record from KVStore.
func (k Keeper) MoveSellOrderToHistorical(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType,
) error {
	// find active record
	so := k.GetSellOrder(ctx, assetId, assetType)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s: %s", assetType.FriendlyString(), assetId)
	}

	if so.HighestBid == nil {
		// in-case of no bid, check if the order has expired
		if !so.HasExpiredAtCtx(ctx) {
			return errorsmod.Wrapf(gerrc.ErrFailedPrecondition,
				"Sell-Order not yet expired: %s", assetId,
			)
		}
	}

	// remove the active record
	k.DeleteSellOrder(ctx, so.AssetId, so.AssetType)

	// set historical records
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(assetId, so.AssetType)
	bz := store.Get(hSoKey)

	var hSo dymnstypes.HistoricalSellOrders
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &hSo)
	}
	hSo.SellOrders = append(hSo.SellOrders, *so)

	if ignorableErr := hSo.Validate(); ignorableErr != nil {
		k.Logger(ctx).Error(
			"historical sell order validation failed, skip persist this historical record",
			"error", ignorableErr,
		)

		// the historical record is not an important data for the chain to function,
		// so in this case, we just skip persisting the invalid historical record.

		return nil
	}

	k.SetHistoricalSellOrders(ctx, assetId, assetType, hSo)

	var minExpiry int64 = -1
	for _, hSo := range hSo.SellOrders {
		if minExpiry < 0 || hSo.ExpireAt < minExpiry {
			minExpiry = hSo.ExpireAt
		}
	}
	if minExpiry > 0 {
		k.SetMinExpiryHistoricalSellOrder(ctx, assetId, assetType, minExpiry)
	}

	return nil
}

// SetHistoricalSellOrders store the Historical Sell-Orders of the corresponding Dym-Name/Alias into the KVStore.
func (k Keeper) SetHistoricalSellOrders(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType, hSo dymnstypes.HistoricalSellOrders,
) {
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(assetId, assetType)
	bz := k.cdc.MustMarshal(&hSo)
	store.Set(hSoKey, bz)
}

// GetHistoricalSellOrders retrieves Historical Sell-Orders of the corresponding Dym-Name/Alias from the KVStore.
func (k Keeper) GetHistoricalSellOrders(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType,
) []dymnstypes.SellOrder {
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(assetId, assetType)

	bz := store.Get(hSoKey)
	if bz == nil {
		return nil
	}

	var hSo dymnstypes.HistoricalSellOrders
	k.cdc.MustUnmarshal(bz, &hSo)

	return hSo.SellOrders
}

// DeleteHistoricalSellOrders deletes the Historical Sell-Orders of specific Dym-Name/Alias from the KVStore.
func (k Keeper) DeleteHistoricalSellOrders(ctx sdk.Context, assetId string, assetType dymnstypes.AssetType) {
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(assetId, assetType)
	store.Delete(hSoKey)
}

// SetActiveSellOrdersExpiration stores the expiration of the active Sell-Orders records into the KVStore.
func (k Keeper) SetActiveSellOrdersExpiration(ctx sdk.Context,
	so *dymnstypes.ActiveSellOrdersExpiration, assetType dymnstypes.AssetType,
) error {
	so.Sort()

	if err := so.Validate(); err != nil {
		return err
	}

	var key []byte
	switch assetType {
	case dymnstypes.TypeName:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.TypeAlias:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid asset type: " + assetType.FriendlyString())
	}

	// persist record
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(so)
	store.Set(key, bz)
	return nil
}

// GetActiveSellOrdersExpiration retrieves the expiration of the active Sell-Orders records from the KVStore.
func (k Keeper) GetActiveSellOrdersExpiration(ctx sdk.Context,
	assetType dymnstypes.AssetType,
) *dymnstypes.ActiveSellOrdersExpiration {
	store := ctx.KVStore(k.storeKey)

	var key []byte
	switch assetType {
	case dymnstypes.TypeName:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.TypeAlias:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid asset type: " + assetType.FriendlyString())
	}

	var record dymnstypes.ActiveSellOrdersExpiration

	bz := store.Get(key)
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &record)
	}

	if record.Records == nil {
		record.Records = make([]dymnstypes.ActiveSellOrdersExpirationRecord, 0)
	}

	return &record
}

// SetMinExpiryHistoricalSellOrder stores the minimum expiry
// of all historical Sell-Orders by each Dym-Name into the KVStore.
func (k Keeper) SetMinExpiryHistoricalSellOrder(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType, minExpiry int64,
) {
	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.MinExpiryHistoricalSellOrdersKey(assetId, assetType)
	if minExpiry < 1 {
		store.Delete(key)
	} else {
		store.Set(key, sdk.Uint64ToBigEndian(uint64(minExpiry)))
	}
}

// GetMinExpiryHistoricalSellOrder retrieves the minimum expiry
// of all historical Sell-Orders by the Dym-Name/Alias from the KVStore.
func (k Keeper) GetMinExpiryHistoricalSellOrder(ctx sdk.Context,
	assetId string, assetType dymnstypes.AssetType,
) (minExpiry int64, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.MinExpiryHistoricalSellOrdersKey(assetId, assetType)
	bz := store.Get(key)
	if bz != nil {
		minExpiry = int64(sdk.BigEndianToUint64(bz))
		found = true
	}
	return
}
