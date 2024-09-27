package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MaxDymNameContactLength is the maximum length allowed for Dym-Name contact.
	MaxDymNameContactLength = 140

	// MaxConfigSize is the maximum size allowed for number Dym-Name configuration per Dym-Name.
	// This is another layer protects spamming the chain with large data.
	MaxConfigSize = 100

	// MinDymNamePriceStepsCount is the minimum number of price steps required for Dym-Name price.
	MinDymNamePriceStepsCount = 4

	// MinAliasPriceStepsCount is the minimum number of price steps required for Alias price.
	MinAliasPriceStepsCount = 4
)

// MinPriceValue is the minimum value allowed for price configuration.
var MinPriceValue = sdkmath.NewInt(1e18)

const (
	// OpGasPlaceSellOrder is the gas consumed when an asset owner creates a Sell-Order for selling the asset.
	OpGasPlaceSellOrder sdk.Gas = 25_000_000
	// OpGasCloseSellOrder is the gas consumed when asset owner closes the Sell-Order.
	OpGasCloseSellOrder sdk.Gas = 5_000_000
	// OpGasCompleteSellOrder is the gas consumed when Sell-Order participant completes it.
	OpGasCompleteSellOrder sdk.Gas = 5_000_000

	// OpGasPlaceBidOnSellOrder is the gas consumed when a buyer bids on a Sell-Order.
	OpGasPlaceBidOnSellOrder sdk.Gas = 20_000_000

	// OpGasConfig is the gas consumed when Dym-Name controller updating Dym-Name configuration,
	// We charge this high amount of gas for extra permanent data
	// needed to be stored like reverse lookup record.
	// We do not charge this fee on Delete operation.
	OpGasConfig sdk.Gas = 35_000_000

	// OpGasUpdateContact is the gas consumed when Dym-Name controller updating Dym-Name contact.
	// We do not charge this fee on clear Contact operation.
	OpGasUpdateContact sdk.Gas = 1_000_000

	// OpGasPutBuyOrder is the gas consumed when a buyer placing a buy order, offer to buy an asset.
	OpGasPutBuyOrder sdk.Gas = 25_000_000

	// OpGasUpdateBuyOrder is the gas consumed when the buyer who placed the buy order,
	// updating the existing offer.
	OpGasUpdateBuyOrder sdk.Gas = 20_000_000

	// OpGasCloseBuyOrder is the gas consumed when the buyer who placed the buy order, closing it.
	OpGasCloseBuyOrder sdk.Gas = 5_000_000
)

const (
	// DoNotModifyDesc is a constant used in flags to indicate that description field should not be updated
	DoNotModifyDesc = "[do-not-modify]"
)

const (
	BuyOrderIdTypeDymNamePrefix = "10"
	BuyOrderIdTypeAliasPrefix   = "20"
)

const (
	// LimitMaxElementsInApiRequest is the maximum number of elements allowed in a single API request.
	LimitMaxElementsInApiRequest = 100
)
