package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
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
	prefixActiveSellOrdersExpiration // deprecated
	prefixCountBuyOrders
	prefixBuyOrder
	prefixRvlBuyerToBuyOrderIds   // reverse lookup store
	prefixRvlAssetIdToBuyOrderIds // reverse lookup store
	prefixRollAppEip155IdToAliases
	prefixRvlAliasToRollAppEip155Id // reverse lookup store
)

const (
	// partialStoreAssetTypeDymName is a part of the store key prefix for the SellOrder records of Dym-Name
	partialStoreAssetTypeDymName = iota

	// partialStoreAssetTypeAlias is a part of the store key prefix for the SellOrder records of Alias
	partialStoreAssetTypeAlias
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

	// KeyPrefixSellOrder is the key prefix for the active SellOrder records of both type DymName/Alias
	KeyPrefixSellOrder = []byte{prefixSellOrder}

	// KeyPrefixDymNameSellOrder is the key prefix for the active SellOrder records of type DymName
	KeyPrefixDymNameSellOrder = []byte{prefixSellOrder, partialStoreAssetTypeDymName}

	// KeyPrefixAliasSellOrder is the key prefix for the active SellOrder records of type Alias
	KeyPrefixAliasSellOrder = []byte{prefixSellOrder, partialStoreAssetTypeAlias}

	// KeyPrefixBuyOrder is the key prefix for the active BuyOrder records regardless asset type DymName/Alias
	KeyPrefixBuyOrder = []byte{prefixBuyOrder}

	// KeyPrefixRvlBuyerToBuyOrderIds is the key prefix for the reverse lookup for BuyOrder IDs by the buyer
	KeyPrefixRvlBuyerToBuyOrderIds = []byte{prefixRvlBuyerToBuyOrderIds}

	// KeyPrefixRvlDymNameToBuyOrderIds is the key prefix for the reverse lookup for BuyOrder IDs by the DymName
	KeyPrefixRvlDymNameToBuyOrderIds = []byte{prefixRvlAssetIdToBuyOrderIds, partialStoreAssetTypeDymName}

	// KeyPrefixRvlAliasToBuyOrderIds is the key prefix for the reverse lookup for BuyOrder IDs by the Alias
	KeyPrefixRvlAliasToBuyOrderIds = []byte{prefixRvlAssetIdToBuyOrderIds, partialStoreAssetTypeAlias}

	// KeyPrefixRollAppEip155IdToAliases is the key prefix for the Roll-App EIP-155 ID to Alias records
	KeyPrefixRollAppEip155IdToAliases = []byte{prefixRollAppEip155IdToAliases}

	// KeyPrefixRvlAliasToRollAppEip155Id is the key prefix for the reverse lookup for Alias to Roll-App EIP-155 ID records
	KeyPrefixRvlAliasToRollAppEip155Id = []byte{prefixRvlAliasToRollAppEip155Id}
)

// KeyCountBuyOrders is the key for the count of all-time buy orders
var KeyCountBuyOrders = []byte{prefixCountBuyOrders}

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

// SellOrderKey returns a key for the active Sell-Order of the Dym-Name/Alias
func SellOrderKey(assetId string, assetType AssetType) []byte {
	switch assetType {
	case TypeName:
		return append(KeyPrefixDymNameSellOrder, []byte(assetId)...)
	case TypeAlias:
		return append(KeyPrefixAliasSellOrder, []byte(assetId)...)
	default:
		panic("invalid asset type: " + assetType.PrettyName())
	}
}

// BuyOrderKey returns a key for the active Buy-Order of the Dym-Name/Alias
func BuyOrderKey(orderId string) []byte {
	return append(KeyPrefixBuyOrder, []byte(orderId)...)
}

// BuyerToOrderIdsRvlKey returns a key for reverse lookup for Buy-Order IDs by the buyer
func BuyerToOrderIdsRvlKey(bzHexAddr []byte) []byte {
	return append(KeyPrefixRvlBuyerToBuyOrderIds, bzHexAddr...)
}

// DymNameToBuyOrderIdsRvlKey returns a key for reverse lookup for Buy-Order IDs by the Dym-Name
func DymNameToBuyOrderIdsRvlKey(dymName string) []byte {
	return append(KeyPrefixRvlDymNameToBuyOrderIds, []byte(dymName)...)
}

// AliasToBuyOrderIdsRvlKey returns a key for reverse lookup for Buy-Order IDs by the Alias
func AliasToBuyOrderIdsRvlKey(alias string) []byte {
	return append(KeyPrefixRvlAliasToBuyOrderIds, []byte(alias)...)
}

// RollAppIdToAliasesKey returns a key for the Roll-App ID to list of alias records.
// Note: the input RollApp ID must be full-chain-id, not EIP-155 only part.
func RollAppIdToAliasesKey(rollAppId string) []byte {
	eip155 := dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollAppId)
	return append(KeyPrefixRollAppEip155IdToAliases, []byte(eip155)...)
}

// AliasToRollAppEip155IdRvlKey returns a key for reverse lookup for Alias to Roll-App EIP-155 ID records
func AliasToRollAppEip155IdRvlKey(alias string) []byte {
	return append(KeyPrefixRvlAliasToRollAppEip155Id, []byte(alias)...)
}
