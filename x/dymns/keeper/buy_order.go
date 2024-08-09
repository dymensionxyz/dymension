package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// IncreaseBuyOrdersCountAndGet increases the all-time Buy-Order records count and returns the updated value.
func (k Keeper) IncreaseBuyOrdersCountAndGet(ctx sdk.Context) uint64 {
	countFromStore := k.GetCountBuyOrders(ctx)
	newCount := countFromStore + 1

	if newCount < countFromStore {
		panic("overflow")
	}

	k.SetCountBuyOrders(ctx, newCount)

	return newCount
}

// GetCountBuyOrders returns the all-time Buy-Order records count from the KVStore.
// Note: do not use this function. This function should only be used in IncreaseBuyOrdersCountAndGet and test.
func (k Keeper) GetCountBuyOrders(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(dymnstypes.KeyCountBuyOrders)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetCountBuyOrders sets the all-time Buy-Order records count into the KVStore.
// Note: do not use this function. This function should only be used in IncreaseBuyOrdersCountAndGet and test.
func (k Keeper) SetCountBuyOrders(ctx sdk.Context, value uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(dymnstypes.KeyCountBuyOrders, sdk.Uint64ToBigEndian(value))
}

// GetAllBuyOrders returns all Buy-Order records from the KVStore.
// No filter is applied.
func (k Keeper) GetAllBuyOrders(ctx sdk.Context) (list []dymnstypes.BuyOrder) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixBuyOrder)
	defer func() {
		_ = iterator.Close() // nolint: errcheck
	}()

	for ; iterator.Valid(); iterator.Next() {
		var offer dymnstypes.BuyOrder
		k.cdc.MustUnmarshal(iterator.Value(), &offer)
		list = append(list, offer)
	}

	return
}

// GetBuyOrder retrieves the Buy-Order from the KVStore.
func (k Keeper) GetBuyOrder(ctx sdk.Context, orderId string) *dymnstypes.BuyOrder {
	if !dymnstypes.IsValidBuyOrderId(orderId) {
		panic("invalid Buy-Order ID")
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOrderKey(orderId)

	bz := store.Get(offerKey)
	if bz == nil {
		return nil
	}

	var offer dymnstypes.BuyOrder
	k.cdc.MustUnmarshal(bz, &offer)

	return &offer
}

// InsertNewBuyOrder assigns ID and insert new Buy-Order record into the KVStore.
func (k Keeper) InsertNewBuyOrder(ctx sdk.Context, buyOrder dymnstypes.BuyOrder) (dymnstypes.BuyOrder, error) {
	if buyOrder.Id != "" {
		panic("ID of the buy order must be empty")
	}

	count := k.IncreaseBuyOrdersCountAndGet(ctx)
	newOrderId := dymnstypes.CreateBuyOrderId(buyOrder.Type, count)

	existingRecord := k.GetBuyOrder(ctx, newOrderId)
	if existingRecord != nil {
		return buyOrder, errorsmod.Wrapf(
			gerrc.ErrAlreadyExists, "Buy-Order ID already exists: %s", newOrderId,
		)
	}

	buyOrder.Id = newOrderId

	if err := k.SetBuyOrder(ctx, buyOrder); err != nil {
		return buyOrder, err
	}

	return buyOrder, nil
}

// SetBuyOrder stores the Buy-Order into the KVStore.
func (k Keeper) SetBuyOrder(ctx sdk.Context, offer dymnstypes.BuyOrder) error {
	if err := offer.Validate(); err != nil {
		return err
	}

	if len(offer.Params) == 0 {
		offer.Params = nil
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOrderKey(offer.Id)
	bz := k.cdc.MustMarshal(&offer)
	store.Set(offerKey, bz)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeValueBoActionNameSet))

	return nil
}

// DeleteBuyOrder deletes the Buy-Order from the KVStore.
func (k Keeper) DeleteBuyOrder(ctx sdk.Context, orderId string) {
	offer := k.GetBuyOrder(ctx, orderId)
	if offer == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOrderKey(orderId)
	store.Delete(offerKey)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeValueBoActionNameDelete))
}
