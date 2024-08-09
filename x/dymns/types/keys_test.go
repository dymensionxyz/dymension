package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestStorePrefixes(t *testing.T) {
	t.Run("ensure key prefixes are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, []byte{0x01}, KeyPrefixDymName, "do not change it, will break the app")
		require.Equal(t, []byte{0x02}, KeyPrefixRvlDymNamesOwnedByAccount, "do not change it, will break the app")
		require.Equal(t, []byte{0x03}, KeyPrefixRvlConfiguredAddressToDymNamesInclude, "do not change it, will break the app")
		require.Equal(t, []byte{0x04}, KeyPrefixRvlFallbackAddressToDymNamesInclude, "do not change it, will break the app")
		require.Equal(t, []byte{0x05}, KeyPrefixSellOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x05, partialStoreOrderTypeDymName}, KeyPrefixDymNameSellOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x07, partialStoreOrderTypeDymName}, KeyPrefixDymNameHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x08, partialStoreOrderTypeDymName}, KeyPrefixMinExpiryDymNameHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x05, partialStoreOrderTypeAlias}, KeyPrefixAliasSellOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x07, partialStoreOrderTypeAlias}, KeyPrefixAliasHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x08, partialStoreOrderTypeAlias}, KeyPrefixMinExpiryAliasHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x0A}, KeyPrefixBuyOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x0B}, KeyPrefixRvlBuyerToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0C, partialStoreOrderTypeDymName}, KeyPrefixRvlDymNameToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0C, partialStoreOrderTypeAlias}, KeyPrefixRvlAliasToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0D}, KeyPrefixRollAppIdToAliases, "do not change it, will break the app")
		require.Equal(t, []byte{0x0E}, KeyPrefixRvlAliasToRollAppId, "do not change it, will break the app")
	})

	t.Run("ensure keys are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, []byte{0x06, partialStoreOrderTypeDymName}, KeyActiveSellOrdersExpirationOfDymName, "do not change it, will break the app")
		require.Equal(t, []byte{0x06, partialStoreOrderTypeAlias}, KeyActiveSellOrdersExpirationOfAlias, "do not change it, will break the app")
		require.Equal(t, []byte{0x09}, KeyCountBuyOrders, "do not change it, will break the app")
	})

	t.Run("ensure partitioned keys are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, byte(0x00), byte(partialStoreOrderTypeDymName), "do not change it, will break the app")
		require.Equal(t, byte(0x01), byte(partialStoreOrderTypeAlias), "do not change it, will break the app")
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeys(t *testing.T) {
	for _, dymName := range []string{"a", "b", "my-name"} {
		t.Run(dymName, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixDymName, []byte(dymName)...), DymNameKey(dymName))
			require.Equal(t, append(KeyPrefixDymNameSellOrder, []byte(dymName)...), SellOrderKey(dymName, NameOrder))
			require.Equal(t, append(KeyPrefixDymNameHistoricalSellOrders, []byte(dymName)...), HistoricalSellOrdersKey(dymName, NameOrder))
			require.Equal(t, append(KeyPrefixMinExpiryDymNameHistoricalSellOrders, []byte(dymName)...), MinExpiryHistoricalSellOrdersKey(dymName, NameOrder))
			require.Equal(t, append(KeyPrefixRvlDymNameToBuyOrderIds, []byte(dymName)...), DymNameToBuyOrderIdsRvlKey(dymName))
		})
	}

	for _, alias := range []string{"a", "b", "alias"} {
		t.Run(alias, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixAliasSellOrder, []byte(alias)...), SellOrderKey(alias, AliasOrder))
			require.Equal(t, append(KeyPrefixAliasHistoricalSellOrders, []byte(alias)...), HistoricalSellOrdersKey(alias, AliasOrder))
			require.Equal(t, append(KeyPrefixMinExpiryAliasHistoricalSellOrders, []byte(alias)...), MinExpiryHistoricalSellOrdersKey(alias, AliasOrder))
			require.Equal(t, append(KeyPrefixRvlAliasToBuyOrderIds, []byte(alias)...), AliasToBuyOrderIdsRvlKey(alias))
		})
	}

	for _, bech32Address := range []string{
		"dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		"dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
		"dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
	} {
		t.Run(bech32Address, func(t *testing.T) {
			accAddr := sdk.MustAccAddressFromBech32(bech32Address)
			require.Equal(t, append(KeyPrefixRvlDymNamesOwnedByAccount, accAddr.Bytes()...), DymNamesOwnedByAccountRvlKey(accAddr))
			require.Equal(t, append(KeyPrefixRvlConfiguredAddressToDymNamesInclude, []byte(bech32Address)...), ConfiguredAddressToDymNamesIncludeRvlKey(bech32Address))
			require.Equal(t, append(KeyPrefixRvlFallbackAddressToDymNamesInclude, accAddr.Bytes()...), FallbackAddressToDymNamesIncludeRvlKey(FallbackAddress(accAddr)))
			require.Equal(t, append(KeyPrefixRvlBuyerToBuyOrderIds, accAddr.Bytes()...), BuyerToOrderIdsRvlKey(accAddr.Bytes()))
		})
	}

	for _, input := range []string{
		"888",
		"aaa",
		"@@@",
	} {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixBuyOrder, []byte(input)...), BuyOrderKey(input))
			require.Equal(t, append(KeyPrefixRvlAliasToRollAppId, []byte(input)...), AliasToRollAppIdRvlKey(input))
		})
	}

	t.Run("should panics of getting Sell-Order related keys if order type is invalid", func(t *testing.T) {
		require.Panics(t, func() { _ = SellOrderKey("goods", OrderType_OT_UNKNOWN) })
		require.Panics(t, func() { _ = HistoricalSellOrdersKey("goods", OrderType_OT_UNKNOWN) })
		require.Panics(t, func() { _ = MinExpiryHistoricalSellOrdersKey("goods", OrderType_OT_UNKNOWN) })
	})
}
