package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	// MaxDymNameLength is the maximum length allowed for Dym-Name.
	MaxDymNameLength = 20

	// MaxDymNameContactLength is the maximum length allowed for Dym-Name contact.
	MaxDymNameContactLength = 140

	// MaxConfigSize is the maximum size allowed for number Dym-Name configuration per Dym-Name.
	// This is another layer protects spamming the chain with large data.
	MaxConfigSize = 100
)

const (
	// OpGasPlaceSellOrder is the gas consumed when a Dym-Name owner creates a Sell-Order for selling Dym-Name.
	OpGasPlaceSellOrder sdk.Gas = 25_000_000
	// OpGasCloseSellOrder is the gas consumed when Dym-Name owner closes Sell-Order.
	OpGasCloseSellOrder sdk.Gas = 5_000_000

	// OpGasPlaceBidOnSellOrder is the gas consumed when a buyer bids on a Dym-Name Sell-Order.
	OpGasPlaceBidOnSellOrder sdk.Gas = 20_000_000

	// OpGasConfig is the gas consumed when Dym-Name controller updating Dym-Name configuration,
	// We charge this high amount of gas for extra permanent data
	// needed to be stored like reverse lookup record.
	// We do not charge this fee on Delete operation.
	OpGasConfig sdk.Gas = 35_000_000

	// OpGasUpdateContact is the gas consumed when Dym-Name controller updating Dym-Name contact.
	// We do not charge this fee on clear Contact operation.
	OpGasUpdateContact sdk.Gas = 1_000_000

	// OpGasPutBuyOffer is the gas consumed when a buyer placing an offer to buy Dym-Name.
	OpGasPutBuyOffer sdk.Gas = 25_000_000

	// OpGasUpdateBuyOffer is the gas consumed when the buyer who placed the buy offer,
	// updating the offer to buy Dym-Name.
	OpGasUpdateBuyOffer sdk.Gas = 20_000_000

	// OpGasCloseBuyOffer is the gas consumed when the buyer who placed the buy offer,
	// closing the offer to buy Dym-Name.
	OpGasCloseBuyOffer sdk.Gas = 5_000_000
)

const (
	// DoNotModifyDesc is a constant used in flags to indicate that description field should not be updated
	DoNotModifyDesc = "[do-not-modify]"
)

const (
	BuyOfferIdTypeDymNamePrefix = "10"
	BuyOfferIdTypeAliasPrefix   = "20"
)
