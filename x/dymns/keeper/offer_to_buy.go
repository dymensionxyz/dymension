package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// IncreaseOfferToBuyCountAndGet increases the count of the Offer-To-Buy and returns the updated value.
func (k Keeper) IncreaseOfferToBuyCountAndGet(ctx sdk.Context) uint64 {
	countFromStore := k.GetCountOfferToBuy(ctx)
	newCount := countFromStore + 1

	if newCount < countFromStore {
		panic("overflow")
	}

	k.SetCountOfferToBuy(ctx, newCount)

	return newCount
}

// GetCountOfferToBuy returns the count of the Offer-To-Buy from the KVStore.
// Note: do not use this function. This function should only be used in IncreaseOfferToBuyCountAndGet and test.
func (k Keeper) GetCountOfferToBuy(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(dymnstypes.KeyCountOfferToBuy)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetCountOfferToBuy sets the count of the Offer-To-Buy into the KVStore.
// Note: do not use this function. This function should only be used in IncreaseOfferToBuyCountAndGet and test.
func (k Keeper) SetCountOfferToBuy(ctx sdk.Context, value uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(dymnstypes.KeyCountOfferToBuy, sdk.Uint64ToBigEndian(value))
}

// GetAllOffersToBuy returns all Offer-To-Buy from the KVStore.
// No filter is applied.
func (k Keeper) GetAllOffersToBuy(ctx sdk.Context) (list []dymnstypes.OfferToBuy) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixOfferToBuy)
	defer func() {
		_ = iterator.Close() // nolint: errcheck
	}()

	for ; iterator.Valid(); iterator.Next() {
		var offer dymnstypes.OfferToBuy
		k.cdc.MustUnmarshal(iterator.Value(), &offer)
		list = append(list, offer)
	}

	return list
}

// GetOfferToBuy retrieves the Offer-To-Buy from the KVStore.
func (k Keeper) GetOfferToBuy(ctx sdk.Context, offerId string) *dymnstypes.OfferToBuy {
	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.OfferToBuyKey(offerId)

	bz := store.Get(offerKey)
	if bz == nil {
		return nil
	}

	var offer dymnstypes.OfferToBuy
	k.cdc.MustUnmarshal(bz, &offer)

	return &offer
}

// InsertOfferToBuy assigns ID and insert new Offer-To-Buy record into the KVStore.
func (k Keeper) InsertOfferToBuy(ctx sdk.Context, offer dymnstypes.OfferToBuy) (dymnstypes.OfferToBuy, error) {
	if offer.Id != "" {
		panic("ID of offer must be empty")
	}

	count := k.IncreaseOfferToBuyCountAndGet(ctx)
	newOfferId := sdkmath.NewIntFromUint64(count).String()

	existingRecord := k.GetOfferToBuy(ctx, newOfferId)
	if existingRecord != nil {
		return offer, sdkerrors.ErrConflict.Wrapf("Offer-To-Buy with ID %s already exists", newOfferId)
	}

	offer.Id = newOfferId

	if err := k.SetOfferToBuy(ctx, offer); err != nil {
		return offer, err
	}

	return offer, nil
}

// SetOfferToBuy stores the Offer-To-Buy into the KVStore.
func (k Keeper) SetOfferToBuy(ctx sdk.Context, offer dymnstypes.OfferToBuy) error {
	if err := offer.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.OfferToBuyKey(offer.Id)
	bz := k.cdc.MustMarshal(&offer)
	store.Set(offerKey, bz)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeKeyOtbActionNameSet))

	return nil
}

// DeleteOfferToBuy deletes the Offer-To-Buy from the KVStore.
func (k Keeper) DeleteOfferToBuy(ctx sdk.Context, offerId string) {
	offer := k.GetOfferToBuy(ctx, offerId)
	if offer == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.OfferToBuyKey(offerId)
	store.Delete(offerKey)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeKeyOtbActionNameDelete))
}
