package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// AddReverseMappingBuyerToOfferToBuyRecord stores a reverse mapping from buyer to Offer-To-Buy Id into the KVStore.
func (k Keeper) AddReverseMappingBuyerToOfferToBuyRecord(ctx sdk.Context, buyer, offerId string) error {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(buyer)
	}

	if !dymnsutils.IsValidBuyNameOfferId(offerId) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Offer-To-Buy Id")
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	return k.GenericAddReverseLookupOfferToBuyIdsRecord(ctx, key, offerId)
}

// GetOfferToBuyByBuyer returns all Offer-To-Buy placed by the account address.
func (k Keeper) GetOfferToBuyByBuyer(
	ctx sdk.Context, buyer string,
) ([]dymnstypes.OfferToBuy, error) {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(buyer)
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	existingOfferIds := k.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)

	var offers []dymnstypes.OfferToBuy
	for _, offerId := range existingOfferIds.OfferIds {
		offer := k.GetOfferToBuy(ctx, offerId)
		if offer == nil {
			// offer not found, skip
			continue
		}
		if offer.Buyer != buyer {
			// offer buyer mismatch, skip
			continue
		}
		offers = append(offers, *offer)
	}

	return offers, nil
}

// RemoveReverseMappingBuyerToOfferToBuy removes a reverse mapping from buyer to Offer-To-Buy Id from the KVStore.
func (k Keeper) RemoveReverseMappingBuyerToOfferToBuy(ctx sdk.Context, buyer, offerId string) error {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap(buyer)
	}

	if !dymnsutils.IsValidBuyNameOfferId(offerId) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Offer-To-Buy Id")
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	return k.GenericRemoveReverseLookupOfferToBuyIdsRecord(ctx, key, offerId)
}

// AddReverseMappingDymNameToOfferToBuy stores a reverse mapping from configured address to Dym-Name which contains the configuration, into the KVStore.
func (k Keeper) AddReverseMappingDymNameToOfferToBuy(ctx sdk.Context, name, offerId string) error {
	if !dymnsutils.IsValidDymName(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Dym-Name")
	}

	if !dymnsutils.IsValidBuyNameOfferId(offerId) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Offer-To-Buy Id")
	}

	return k.GenericAddReverseLookupOfferToBuyIdsRecord(
		ctx,
		dymnstypes.DymNameToOfferIdsRvlKey(name),
		offerId,
	)
}

// GetOffersToBuyOfDymName returns all Offers-To-Buy that placed for the Dym-Name.
func (k Keeper) GetOffersToBuyOfDymName(
	ctx sdk.Context, name string,
) ([]dymnstypes.OfferToBuy, error) {
	if !dymnsutils.IsValidDymName(name) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("invalid Dym-Name")
	}

	key := dymnstypes.DymNameToOfferIdsRvlKey(name)

	offerIds := k.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)

	var offers []dymnstypes.OfferToBuy
	for _, offerId := range offerIds.OfferIds {
		offer := k.GetOfferToBuy(ctx, offerId)
		if offer == nil {
			// offer not found, skip
			continue
		}
		offers = append(offers, *offer)
	}

	return offers, nil
}

// RemoveReverseMappingDymNameToOfferToBuy removes reverse mapping from Dym-Name to Offer-To-Buy which placed for it, from the KVStore.
func (k Keeper) RemoveReverseMappingDymNameToOfferToBuy(ctx sdk.Context, name, offerId string) error {
	if !dymnsutils.IsValidDymName(name) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Dym-Name")
	}

	if !dymnsutils.IsValidBuyNameOfferId(offerId) {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid Offer-To-Buy Id")
	}

	return k.GenericRemoveReverseLookupOfferToBuyIdsRecord(
		ctx,
		dymnstypes.DymNameToOfferIdsRvlKey(name),
		offerId,
	)
}

// GenericAddReverseLookupOfferToBuyIdsRecord is a utility method that help to add a reverse lookup record for Offer-To-Buy Ids.
func (k Keeper) GenericAddReverseLookupOfferToBuyIdsRecord(ctx sdk.Context, key []byte, offerId string) error {
	return k.GenericAddReverseLookupRecord(
		ctx,
		key, offerId,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupOfferIds{
				OfferIds: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupOfferIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OfferIds
		},
	)
}

// GenericGetReverseLookupOfferToBuyIdsRecord is a utility method that help to get a reverse lookup record for Offer-To-Buy Ids.
func (k Keeper) GenericGetReverseLookupOfferToBuyIdsRecord(
	ctx sdk.Context, key []byte,
) dymnstypes.ReverseLookupOfferIds {
	dymNames := k.GenericGetReverseLookupRecord(
		ctx,
		key,
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupOfferIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OfferIds
		},
	)

	return dymnstypes.ReverseLookupOfferIds{
		OfferIds: dymNames,
	}
}

// GenericRemoveReverseLookupOfferToBuyIdsRecord is a utility method that help to remove a reverse lookup record for Offer-To-Buy Ids.
func (k Keeper) GenericRemoveReverseLookupOfferToBuyIdsRecord(ctx sdk.Context, key []byte, offerId string) error {
	return k.GenericRemoveReverseLookupRecord(
		ctx,
		key, offerId,
		func(list []string) []byte {
			record := dymnstypes.ReverseLookupOfferIds{
				OfferIds: list,
			}
			return k.cdc.MustMarshal(&record)
		},
		func(bz []byte) []string {
			var record dymnstypes.ReverseLookupOfferIds
			k.cdc.MustUnmarshal(bz, &record)
			return record.OfferIds
		},
	)
}
