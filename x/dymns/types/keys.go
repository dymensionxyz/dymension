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
	prefixRvlFallbackAddressToDymNamesInclude   // reverse lookup store
	prefixSellOrder
	prefixActiveSellOrdersExpiration
	prefixHistoricalSellOrders
	prefixMinExpiryHistoricalSellOrders
	prefixCountBuyOffers
	prefixBuyOffer
	prefixRvlBuyerToOfferIds   // reverse lookup store
	prefixRvlDymNameToOfferIds // reverse lookup store
	prefixRvlAliasToOfferIds   // reverse lookup store
	prefixRollAppIdToAlias
	prefixRvlAliasToRollAppId // reverse lookup store
)

const (
	// partialStoreOrderTypeDymName is a part of the store key prefix for the SellOrder records of Dym-Name
	partialStoreOrderTypeDymName = iota

	// partialStoreOrderTypeAlias is a part of the store key prefix for the SellOrder records of Alias
	partialStoreOrderTypeAlias
)

var (
	// KeyPrefixDymName is the key prefix for the DymName records
	KeyPrefixDymName = []byte{prefixDymName}

	// KeyPrefixRvlDymNamesOwnedByAccount is the key prefix for the reverse lookup for Dym-Names owned by an account
	KeyPrefixRvlDymNamesOwnedByAccount = []byte{prefixRvlDymNamesOwnedByAccount}

	// KeyPrefixRvlConfiguredAddressToDymNamesInclude is the key prefix for the reverse lookup for Dym-Names that contain the configured address (bech32)
	KeyPrefixRvlConfiguredAddressToDymNamesInclude = []byte{prefixRvlConfiguredAddressToDymNamesInclude}

	// KeyPrefixRvlFallbackAddressToDymNamesInclude is the key prefix for the reverse lookup address for Dym-Names using fallback mechanism
	KeyPrefixRvlFallbackAddressToDymNamesInclude = []byte{prefixRvlFallbackAddressToDymNamesInclude}

	// KeyPrefixDymNameSellOrder is the key prefix for the active SellOrder records of type DymName
	KeyPrefixDymNameSellOrder = []byte{prefixSellOrder, partialStoreOrderTypeDymName}

	// KeyPrefixDymNameHistoricalSellOrders is the key prefix for the historical SellOrder records of type DymName
	KeyPrefixDymNameHistoricalSellOrders = []byte{prefixHistoricalSellOrders, partialStoreOrderTypeDymName}

	// KeyPrefixMinExpiryDymNameHistoricalSellOrders is the key prefix for the lowest expiry among the historical SellOrder records of each specific Dym-Name
	KeyPrefixMinExpiryDymNameHistoricalSellOrders = []byte{prefixMinExpiryHistoricalSellOrders, partialStoreOrderTypeDymName}

	// KeyPrefixBuyOrder is the key prefix for the active BuyOffer records regardless order type DymName/Alias
	KeyPrefixBuyOrder = []byte{prefixBuyOffer}

	// KeyPrefixRvlBuyerToOfferIds is the key prefix for the reverse lookup for BuyOffer IDs by the buyer
	KeyPrefixRvlBuyerToOfferIds = []byte{prefixRvlBuyerToOfferIds}

	// KeyPrefixRvlDymNameToOfferIds is the key prefix for the reverse lookup for BuyOffer IDs by the DymName
	KeyPrefixRvlDymNameToOfferIds = []byte{prefixRvlDymNameToOfferIds}

	// KeyPrefixRvlAliasToOfferIds is the key prefix for the reverse lookup for BuyOffer IDs by the Alias
	KeyPrefixRvlAliasToOfferIds = []byte{prefixRvlAliasToOfferIds}

	// KeyPrefixRollAppIdToAlias is the key prefix for the Roll-App ID to Alias records
	KeyPrefixRollAppIdToAlias = []byte{prefixRollAppIdToAlias}

	// KeyPrefixRvlAliasToRollAppId is the key prefix for the reverse lookup for Alias to Roll-App ID records
	KeyPrefixRvlAliasToRollAppId = []byte{prefixRvlAliasToRollAppId}
)

var (
	KeyActiveSellOrdersExpiration = []byte{prefixActiveSellOrdersExpiration}

	// KeyCountBuyOffers is the key for the count of all-time buy offer orders
	KeyCountBuyOffers = []byte{prefixCountBuyOffers}
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

// FallbackAddressToDymNamesIncludeRvlKey returns the key for the reverse lookup address for Dym-Names using fallback mechanism
func FallbackAddressToDymNamesIncludeRvlKey(fallbackAddr FallbackAddress) []byte {
	return append(KeyPrefixRvlFallbackAddressToDymNamesInclude, fallbackAddr...)
}

// SellOrderKey returns a key for the active Sell-Order of the Dym-Name
func SellOrderKey(dymName string) []byte {
	return append(KeyPrefixDymNameSellOrder, []byte(dymName)...)
}

// HistoricalSellOrdersKey returns a key for the historical Sell-Orders of the Dym-Name
func HistoricalSellOrdersKey(dymName string) []byte {
	return append(KeyPrefixDymNameHistoricalSellOrders, []byte(dymName)...)
}

// MinExpiryHistoricalSellOrdersKey returns a key for lowest expiry among the historical Sell-Orders
// of the Dym-Name
func MinExpiryHistoricalSellOrdersKey(dymName string) []byte {
	return append(KeyPrefixMinExpiryDymNameHistoricalSellOrders, []byte(dymName)...)
}

// BuyOfferKey returns a key for the active Buy-Order of the Dym-Name
func BuyOfferKey(offerId string) []byte {
	return append(KeyPrefixBuyOrder, []byte(offerId)...)
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
