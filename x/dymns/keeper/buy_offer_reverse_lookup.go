package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// AddReverseMappingBuyerToBuyOfferRecord stores a reverse mapping from buyer to IDs of Buy-Order into the KVStore.
func (k Keeper) AddReverseMappingBuyerToBuyOfferRecord(ctx sdk.Context, buyer, offerId string) error {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	if !dymnstypes.IsValidBuyOfferId(offerId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Offer ID: %s", offerId)
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	return k.GenericAddReverseLookupBuyOfferIdsRecord(ctx, key, offerId)
}

// GetBuyOffersByBuyer returns all Buy-Orders placed by the account address.
func (k Keeper) GetBuyOffersByBuyer(
	ctx sdk.Context, buyer string,
) ([]dymnstypes.BuyOffer, error) {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	existingOfferIds := k.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)

	var offers []dymnstypes.BuyOffer
	for _, offerId := range existingOfferIds.OfferIds {
		offer := k.GetBuyOffer(ctx, offerId)
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

// RemoveReverseMappingBuyerToBuyOffer removes a reverse mapping from buyer to a Buy-Order ID from the KVStore.
func (k Keeper) RemoveReverseMappingBuyerToBuyOffer(ctx sdk.Context, buyer, offerId string) error {
	_, bzAccAddr, err := bech32.DecodeAndConvert(buyer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid buyer address: %s", buyer)
	}

	if !dymnstypes.IsValidBuyOfferId(offerId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Offer ID: %s", offerId)
	}

	key := dymnstypes.BuyerToOfferIdsRvlKey(bzAccAddr)

	return k.GenericRemoveReverseLookupBuyOfferIdRecord(ctx, key, offerId)
}

// AddReverseMappingGoodsIdToBuyOffer add a reverse mapping from Dym-Name/Alias to the Buy-Order ID which placed for it, into the KVStore.
func (k Keeper) AddReverseMappingGoodsIdToBuyOffer(ctx sdk.Context, goodsId string, orderType dymnstypes.OrderType, offerId string) error {
	var key []byte

	switch orderType {
	case dymnstypes.NameOrder:
		if !dymnsutils.IsValidDymName(goodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", goodsId)
		}
		key = dymnstypes.DymNameToOfferIdsRvlKey(goodsId)
	case dymnstypes.AliasOrder:
		if !dymnsutils.IsValidAlias(goodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", goodsId)
		}
		key = dymnstypes.AliasToOfferIdsRvlKey(goodsId)
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", orderType)
	}

	if !dymnstypes.IsValidBuyOfferId(offerId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Offer ID: %s", offerId)
	}

	return k.GenericAddReverseLookupBuyOfferIdsRecord(ctx, key, offerId)
}

// GetBuyOffersOfDymName returns all Buy-Orders that placed for the Dym-Name.
func (k Keeper) GetBuyOffersOfDymName(
	ctx sdk.Context, name string,
) ([]dymnstypes.BuyOffer, error) {
	if !dymnsutils.IsValidDymName(name) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", name)
	}

	key := dymnstypes.DymNameToOfferIdsRvlKey(name)

	offerIds := k.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)

	var offers []dymnstypes.BuyOffer
	for _, offerId := range offerIds.OfferIds {
		offer := k.GetBuyOffer(ctx, offerId)
		if offer == nil {
			// offer not found, skip
			continue
		}
		offers = append(offers, *offer)
	}

	return offers, nil
}

// GetBuyOffersOfAlias returns all Buy-Orders that placed for the Alias.
func (k Keeper) GetBuyOffersOfAlias(
	ctx sdk.Context, alias string,
) ([]dymnstypes.BuyOffer, error) {
	if !dymnsutils.IsValidAlias(alias) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", alias)
	}

	key := dymnstypes.AliasToOfferIdsRvlKey(alias)

	offerIds := k.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)

	var offers []dymnstypes.BuyOffer
	for _, offerId := range offerIds.OfferIds {
		offer := k.GetBuyOffer(ctx, offerId)
		if offer == nil {
			// offer not found, skip
			continue
		}
		offers = append(offers, *offer)
	}

	return offers, nil
}

// RemoveReverseMappingGoodsIdToBuyOffer removes reverse mapping from Dym-Name/Alias to Buy-Order which placed for it, from the KVStore.
func (k Keeper) RemoveReverseMappingGoodsIdToBuyOffer(ctx sdk.Context, goodsId string, orderType dymnstypes.OrderType, offerId string) error {
	var key []byte

	switch orderType {
	case dymnstypes.NameOrder:
		if !dymnsutils.IsValidDymName(goodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Dym-Name: %s", goodsId)
		}
		key = dymnstypes.DymNameToOfferIdsRvlKey(goodsId)
	case dymnstypes.AliasOrder:
		if !dymnsutils.IsValidAlias(goodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Alias: %s", goodsId)
		}
		key = dymnstypes.AliasToOfferIdsRvlKey(goodsId)
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", orderType)
	}

	if !dymnstypes.IsValidBuyOfferId(offerId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid Buy-Offer ID: %s", offerId)
	}

	return k.GenericRemoveReverseLookupBuyOfferIdRecord(ctx, key, offerId)
}

// GenericAddReverseLookupBuyOfferIdsRecord is a utility method that help to add a reverse lookup record for Buy-Order ID.
func (k Keeper) GenericAddReverseLookupBuyOfferIdsRecord(ctx sdk.Context, key []byte, offerId string) error {
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

// GenericGetReverseLookupBuyOfferIdsRecord is a utility method that help to get a reverse lookup record for Buy-Order IDs.
func (k Keeper) GenericGetReverseLookupBuyOfferIdsRecord(
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

// GenericRemoveReverseLookupBuyOfferIdRecord is a utility method that help to remove a reverse lookup record for Buy-Order ID.
func (k Keeper) GenericRemoveReverseLookupBuyOfferIdRecord(ctx sdk.Context, key []byte, offerId string) error {
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
