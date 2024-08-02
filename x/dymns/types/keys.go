package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "dymns"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// prefix bytes for the DymNS persistent store.
const (
	prefixDymName                               = iota + 1
	prefixRvlDymNamesOwnedByAccount             // reverse lookup store
	prefixRvlConfiguredAddressToDymNamesInclude // reverse lookup store
	prefixRvlHexAddressToDymNamesInclude        // reverse lookup store
	prefixSellOrder
	prefixActiveSellOrdersExpiration
	prefixHistoricalSellOrders
	prefixMinExpiryHistoricalSellOrders
	prefixCountOfferToBuy
	prefixOfferToBuy
	prefixRvlBuyerToOfferIds   // reverse lookup store
	prefixRvlDymNameToOfferIds // reverse lookup store
	prefixRollAppIdToAlias
	prefixRvlAliasToRollAppId // reverse lookup store
)

var (
	// KeyPrefixDymName is the key prefix for the Dym-Name records
	KeyPrefixDymName = []byte{prefixDymName}

	// KeyPrefixRvlDymNamesOwnedByAccount is the key prefix for the reverse lookup for Dym-Names owned by an account
	KeyPrefixRvlDymNamesOwnedByAccount = []byte{prefixRvlDymNamesOwnedByAccount}

	// KeyPrefixRvlConfiguredAddressToDymNamesInclude is the key prefix for the reverse lookup for Dym-Names that contain the configured address (bech32)
	KeyPrefixRvlConfiguredAddressToDymNamesInclude = []byte{prefixRvlConfiguredAddressToDymNamesInclude}

	// KeyPrefixRvlHexAddressToDymNamesInclude is the key prefix for the reverse lookup for Dym-Names that contain the hex address (coin-type 60, secp256k1, ethereum address)
	KeyPrefixRvlHexAddressToDymNamesInclude = []byte{prefixRvlHexAddressToDymNamesInclude}

	// KeyPrefixSellOrder is the key prefix for the active Sell-Order records
	KeyPrefixSellOrder = []byte{prefixSellOrder}

	// KeyPrefixHistoricalSellOrders is the key prefix for the historical Sell-Order records
	KeyPrefixHistoricalSellOrders = []byte{prefixHistoricalSellOrders}

	// KeyPrefixMinExpiryHistoricalSellOrders is the key prefix for the lowest expiry among the historical Sell-Order records of each specific Dym-Name
	KeyPrefixMinExpiryHistoricalSellOrders = []byte{prefixMinExpiryHistoricalSellOrders}

	// KeyPrefixOfferToBuy is the key prefix for the active Offer-To-Buy records
	KeyPrefixOfferToBuy = []byte{prefixOfferToBuy}

	// KeyPrefixRvlBuyerToOfferIds is the key prefix for the reverse lookup for Buy-Order IDs by the buyer
	KeyPrefixRvlBuyerToOfferIds = []byte{prefixRvlBuyerToOfferIds}

	// KeyPrefixRvlDymNameToOfferIds is the key prefix for the reverse lookup for Buy-Order IDs by the Dym-Name
	KeyPrefixRvlDymNameToOfferIds = []byte{prefixRvlDymNameToOfferIds}

	// KeyPrefixRollAppIdToAlias is the key prefix for the Roll-App ID to Alias records
	KeyPrefixRollAppIdToAlias = []byte{prefixRollAppIdToAlias}

	// KeyPrefixRvlAliasToRollAppId is the key prefix for the reverse lookup for Alias to Roll-App ID records
	KeyPrefixRvlAliasToRollAppId = []byte{prefixRvlAliasToRollAppId}
)

var (
	KeyActiveSellOrdersExpiration = []byte{prefixActiveSellOrdersExpiration}

	// KeyCountOfferToBuy is the key for the count of all-time Offer-To-Buy orders
	KeyCountOfferToBuy = []byte{prefixCountOfferToBuy}
)

// DymNameKey returns a key for specific Dym-Name
func DymNameKey(name string) []byte {
	return append(KeyPrefixDymName, []byte(name)...)
}

// DymNamesOwnedByAccountRvlKey returns a key for reverse lookup for Dym-Names owned by an account
func DymNamesOwnedByAccountRvlKey(owner sdk.AccAddress) []byte {
	return append(KeyPrefixRvlDymNamesOwnedByAccount, owner.Bytes()...)
}

// ConfiguredAddressToDymNamesIncludeRvlKey returns a key for reverse lookup for Dym-Names that contain the configured address
func ConfiguredAddressToDymNamesIncludeRvlKey(address string) []byte {
	return append(KeyPrefixRvlConfiguredAddressToDymNamesInclude, []byte(address)...)
}

// HexAddressToDymNamesIncludeRvlKey returns a key for reverse lookup for Dym-Names that contain the hex address (coin-type 60, secp256k1, ethereum address)
func HexAddressToDymNamesIncludeRvlKey(bzHexAddr []byte) []byte {
	return append(KeyPrefixRvlHexAddressToDymNamesInclude, bzHexAddr...)
}

// SellOrderKey returns a key for the active Sell-Order of the Dym-Name
func SellOrderKey(dymName string) []byte {
	return append(KeyPrefixSellOrder, []byte(dymName)...)
}

// HistoricalSellOrdersKey returns a key for the historical Sell-Orders of the Dym-Name
func HistoricalSellOrdersKey(dymName string) []byte {
	return append(KeyPrefixHistoricalSellOrders, []byte(dymName)...)
}

// MinExpiryHistoricalSellOrdersKey returns a key for lowest expiry among the historical Sell-Orders
// of the Dym-Name
func MinExpiryHistoricalSellOrdersKey(dymName string) []byte {
	return append(KeyPrefixMinExpiryHistoricalSellOrders, []byte(dymName)...)
}

// OfferToBuyKey returns a key for the active Buy-Order of the Dym-Name
func OfferToBuyKey(offerId string) []byte {
	return append(KeyPrefixOfferToBuy, []byte(offerId)...)
}

// BuyerToOfferIdsRvlKey returns a key for reverse lookup for Buy-Order IDs by the buyer
func BuyerToOfferIdsRvlKey(bzHexAddr []byte) []byte {
	return append(KeyPrefixRvlBuyerToOfferIds, bzHexAddr...)
}

// DymNameToOfferIdsRvlKey returns a key for reverse lookup for Buy-Order IDs by the Dym-Name
func DymNameToOfferIdsRvlKey(dymName string) []byte {
	return append(KeyPrefixRvlDymNameToOfferIds, []byte(dymName)...)
}

// RollAppIdToAliasKey returns a key for the Roll-App ID to Alias records
func RollAppIdToAliasKey(rollAppId string) []byte {
	return append(KeyPrefixRollAppIdToAlias, []byte(rollAppId)...)
}

// AliasToRollAppIdRvlKey returns a key for reverse lookup for Alias to Roll-App ID records
func AliasToRollAppIdRvlKey(alias string) []byte {
	return append(KeyPrefixRvlAliasToRollAppId, []byte(alias)...)
}
