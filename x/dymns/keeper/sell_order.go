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

	// persist record
	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(so.GoodsId, so.Type)
	bz := k.cdc.MustMarshal(&so)
	store.Set(soKey, bz)

	ctx.EventManager().EmitEvent(so.GetSdkEvent(dymnstypes.AttributeValueSoActionNameSet))

	return nil
}

// GetSellOrder retrieves active Sell-Order of the corresponding Dym-Name/Alias from the KVStore.
// If the Sell-Order does not exist, nil is returned.
func (k Keeper) GetSellOrder(ctx sdk.Context,
	goodsId string, orderType dymnstypes.OrderType,
) *dymnstypes.SellOrder {
	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(goodsId, orderType)

	bz := store.Get(soKey)
	if bz == nil {
		return nil
	}

	var so dymnstypes.SellOrder
	k.cdc.MustUnmarshal(bz, &so)

	return &so
}

// DeleteSellOrder deletes the Sell-Order from the KVStore.
func (k Keeper) DeleteSellOrder(ctx sdk.Context, goodsId string, orderType dymnstypes.OrderType) {
	so := k.GetSellOrder(ctx, goodsId, orderType)
	if so == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	soKey := dymnstypes.SellOrderKey(goodsId, orderType)
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
	goodsId string, orderType dymnstypes.OrderType,
) error {
	// find active record
	so := k.GetSellOrder(ctx, goodsId, orderType)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s: %s", orderType.FriendlyString(), goodsId)
	}

	if so.HighestBid == nil {
		// in-case of no bid, check if the order has expired
		if !so.HasExpiredAtCtx(ctx) {
			return errorsmod.Wrapf(gerrc.ErrFailedPrecondition,
				"Sell-Order not yet expired: %s", goodsId,
			)
		}
	}

	// remove the active record
	k.DeleteSellOrder(ctx, so.GoodsId, so.Type)

	// set historical records
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(goodsId, so.Type)
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

	k.SetHistoricalSellOrders(ctx, goodsId, orderType, hSo)

	var minExpiry int64 = -1
	for _, hSo := range hSo.SellOrders {
		if minExpiry < 0 || hSo.ExpireAt < minExpiry {
			minExpiry = hSo.ExpireAt
		}
	}
	if minExpiry > 0 {
		k.SetMinExpiryHistoricalSellOrder(ctx, goodsId, orderType, minExpiry)
	}

	return nil
}

// SetHistoricalSellOrders store the Historical Sell-Orders of the corresponding Dym-Name/Alias into the KVStore.
func (k Keeper) SetHistoricalSellOrders(ctx sdk.Context,
	goodsId string, orderType dymnstypes.OrderType, hSo dymnstypes.HistoricalSellOrders) {

	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(goodsId, orderType)
	bz := k.cdc.MustMarshal(&hSo)
	store.Set(hSoKey, bz)
}

// GetHistoricalSellOrders retrieves Historical Sell-Orders of the corresponding Dym-Name/Alias from the KVStore.
func (k Keeper) GetHistoricalSellOrders(ctx sdk.Context,
	goodsId string, orderType dymnstypes.OrderType,
) []dymnstypes.SellOrder {
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(goodsId, orderType)

	bz := store.Get(hSoKey)
	if bz == nil {
		return nil
	}

	var hSo dymnstypes.HistoricalSellOrders
	k.cdc.MustUnmarshal(bz, &hSo)

	return hSo.SellOrders
}

// DeleteHistoricalSellOrders deletes the Historical Sell-Orders of specific Dym-Name/Alias from the KVStore.
func (k Keeper) DeleteHistoricalSellOrders(ctx sdk.Context, goodsId string, orderType dymnstypes.OrderType) {
	store := ctx.KVStore(k.storeKey)
	hSoKey := dymnstypes.HistoricalSellOrdersKey(goodsId, orderType)
	store.Delete(hSoKey)
}

// SetActiveSellOrdersExpiration stores the expiration of the active Sell-Orders records into the KVStore.
func (k Keeper) SetActiveSellOrdersExpiration(ctx sdk.Context,
	so *dymnstypes.ActiveSellOrdersExpiration, orderType dymnstypes.OrderType,
) error {
	so.Sort()

	if err := so.Validate(); err != nil {
		return err
	}

	var key []byte
	switch orderType {
	case dymnstypes.NameOrder:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.AliasOrder:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid order type: " + orderType.FriendlyString())
	}

	// persist record
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(so)
	store.Set(key, bz)
	return nil
}

// GetActiveSellOrdersExpiration retrieves the expiration of the active Sell-Orders records from the KVStore.
func (k Keeper) GetActiveSellOrdersExpiration(ctx sdk.Context,
	orderType dymnstypes.OrderType,
) *dymnstypes.ActiveSellOrdersExpiration {
	store := ctx.KVStore(k.storeKey)

	var key []byte
	switch orderType {
	case dymnstypes.NameOrder:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfDymName
	case dymnstypes.AliasOrder:
		key = dymnstypes.KeyActiveSellOrdersExpirationOfAlias
	default:
		panic("invalid order type: " + orderType.FriendlyString())
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
	goodsId string, orderType dymnstypes.OrderType, minExpiry int64,
) {
	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.MinExpiryHistoricalSellOrdersKey(goodsId, orderType)
	if minExpiry < 1 {
		store.Delete(key)
	} else {
		store.Set(key, sdk.Uint64ToBigEndian(uint64(minExpiry)))
	}
}

// GetMinExpiryHistoricalSellOrder retrieves the minimum expiry
// of all historical Sell-Orders by the Dym-Name/Alias from the KVStore.
func (k Keeper) GetMinExpiryHistoricalSellOrder(ctx sdk.Context,
	goodsId string, orderType dymnstypes.OrderType,
) (minExpiry int64, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.MinExpiryHistoricalSellOrdersKey(goodsId, orderType)
	bz := store.Get(key)
	if bz != nil {
		minExpiry = int64(sdk.BigEndianToUint64(bz))
		found = true
	}
	return
}
