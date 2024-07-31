package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	MaxDymNameLength = 20

	MaxDymNameContactLength = 140
)

const (
	// OpGasPutAds is the gas consumed for Dym-Name owner to put an ads for selling Dym-Name.
	OpGasPutAds sdk.Gas = 25_000_000
	// OpGasCloseAds is the gas consumed for Dym-Name owner to close the Dym-Name ads.
	OpGasCloseAds sdk.Gas = 5_000_000

	// OpGasBidAds is the gas consumed for bidding an ads for Dym-Name.
	OpGasBidAds sdk.Gas = 20_000_000

	// OpGasConfig is the gas consumed for updating Dym-Name configuration,
	// We charge this high amount of gas for extra permanent data
	// needed to be stored like reverse lookup record.
	// We do not charge this fee on Delete operation.
	OpGasConfig sdk.Gas = 30_000_000

	// OpGasUpdateContact is the gas consumed for updating Dym-Name contact.
	// We do not charge this fee on clear Contact operation.
	OpGasUpdateContact sdk.Gas = 1_000_000

	// OpGasPutOffer is the gas consumed for putting an offer to buy Dym-Name.
	OpGasPutOffer sdk.Gas = 25_000_000

	// OpGasUpdateOffer is the gas consumed for updating the offer to buy Dym-Name.
	OpGasUpdateOffer sdk.Gas = 20_000_000

	// OpGasCloseOffer is the gas consumed for closing the offer to buy Dym-Name.
	OpGasCloseOffer sdk.Gas = 5_000_000
)

const (
	// DoNotModifyDesc is a constant used in flags to indicate that description field should not be updated
	DoNotModifyDesc = "[do-not-modify]"
)
