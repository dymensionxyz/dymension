package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// IncreaseBuyOfferCountAndGet increases the all-time Buy-Order records count and returns the updated value.
func (k Keeper) IncreaseBuyOfferCountAndGet(ctx sdk.Context) uint64 {
	countFromStore := k.GetCountBuyOffer(ctx)
	newCount := countFromStore + 1

	if newCount < countFromStore {
		panic("overflow")
	}

	k.SetCountBuyOffer(ctx, newCount)

	return newCount
}

// GetCountBuyOffer returns the all-time Buy-Order records count from the KVStore.
// Note: do not use this function. This function should only be used in IncreaseBuyOfferCountAndGet and test.
func (k Keeper) GetCountBuyOffer(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(dymnstypes.KeyCountBuyOffers)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetCountBuyOffer sets the all-time Buy-Order records count into the KVStore.
// Note: do not use this function. This function should only be used in IncreaseBuyOfferCountAndGet and test.
func (k Keeper) SetCountBuyOffer(ctx sdk.Context, value uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(dymnstypes.KeyCountBuyOffers, sdk.Uint64ToBigEndian(value))
}

// GetAllBuyOffers returns all Buy-Order records from the KVStore.
// No filter is applied.
func (k Keeper) GetAllBuyOffers(ctx sdk.Context) (list []dymnstypes.BuyOffer) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixBuyOrder)
	defer func() {
		_ = iterator.Close() // nolint: errcheck
	}()

	for ; iterator.Valid(); iterator.Next() {
		var offer dymnstypes.BuyOffer
		k.cdc.MustUnmarshal(iterator.Value(), &offer)
		list = append(list, offer)
	}

	return list
}

// GetBuyOffer retrieves the Buy-Order from the KVStore.
func (k Keeper) GetBuyOffer(ctx sdk.Context, offerId string) *dymnstypes.BuyOffer {
	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOfferKey(offerId)

	bz := store.Get(offerKey)
	if bz == nil {
		return nil
	}

	var offer dymnstypes.BuyOffer
	k.cdc.MustUnmarshal(bz, &offer)

	return &offer
}

// InsertNewBuyOffer assigns ID and insert new Buy-Order record into the KVStore.
func (k Keeper) InsertNewBuyOffer(ctx sdk.Context, offer dymnstypes.BuyOffer) (dymnstypes.BuyOffer, error) {
	if offer.Id != "" {
		panic("ID of offer must be empty")
	}

	count := k.IncreaseBuyOfferCountAndGet(ctx)
	newOfferId := sdkmath.NewIntFromUint64(count).String()

	existingRecord := k.GetBuyOffer(ctx, newOfferId)
	if existingRecord != nil {
		return offer, errorsmod.Wrapf(
			gerrc.ErrAlreadyExists, "Buy-Order-ID already exists: %s", newOfferId,
		)
	}

	offer.Id = newOfferId

	if err := k.SetBuyOffer(ctx, offer); err != nil {
		return offer, err
	}

	return offer, nil
}

// SetBuyOffer stores the Buy-Order into the KVStore.
func (k Keeper) SetBuyOffer(ctx sdk.Context, offer dymnstypes.BuyOffer) error {
	if err := offer.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOfferKey(offer.Id)
	bz := k.cdc.MustMarshal(&offer)
	store.Set(offerKey, bz)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeValueBoActionNameSet))

	return nil
}

// DeleteBuyOffer deletes the Buy-Order from the KVStore.
func (k Keeper) DeleteBuyOffer(ctx sdk.Context, offerId string) {
	offer := k.GetBuyOffer(ctx, offerId)
	if offer == nil {
		return
	}

	store := ctx.KVStore(k.storeKey)
	offerKey := dymnstypes.BuyOfferKey(offerId)
	store.Delete(offerKey)

	ctx.EventManager().EmitEvent(offer.GetSdkEvent(dymnstypes.AttributeValueBoActionNameDelete))
}
