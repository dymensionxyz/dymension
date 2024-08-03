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
		require.Equal(t, []byte{0x07}, KeyPrefixHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x08}, KeyPrefixMinExpiryHistoricalSellOrders, "do not change it, will break the app")
		require.Equal(t, []byte{0x0A}, KeyPrefixOfferToBuy, "do not change it, will break the app")
		require.Equal(t, []byte{0x0B}, KeyPrefixRvlBuyerToOfferIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0C}, KeyPrefixRvlDymNameToOfferIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0D}, KeyPrefixRollAppIdToAlias, "do not change it, will break the app")
		require.Equal(t, []byte{0x0E}, KeyPrefixRvlAliasToRollAppId, "do not change it, will break the app")
	})

	t.Run("ensure key are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, []byte{0x06}, KeyActiveSellOrdersExpiration, "do not change it, will break the app")
		require.Equal(t, []byte{0x09}, KeyCountOfferToBuy, "do not change it, will break the app")
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeys(t *testing.T) {
	for _, dymName := range []string{"a", "b", "my-name"} {
		t.Run(dymName, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixDymName, []byte(dymName)...), DymNameKey(dymName))
			require.Equal(t, append(KeyPrefixSellOrder, []byte(dymName)...), SellOrderKey(dymName))
			require.Equal(t, append(KeyPrefixHistoricalSellOrders, []byte(dymName)...), HistoricalSellOrdersKey(dymName))
			require.Equal(t, append(KeyPrefixMinExpiryHistoricalSellOrders, []byte(dymName)...), MinExpiryHistoricalSellOrdersKey(dymName))
			require.Equal(t, append(KeyPrefixRvlDymNameToOfferIds, []byte(dymName)...), DymNameToOfferIdsRvlKey(dymName))
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
			require.Equal(t, append(KeyPrefixRvlBuyerToOfferIds, accAddr.Bytes()...), BuyerToOfferIdsRvlKey(accAddr.Bytes()))
		})
	}

	for _, input := range []string{
		"888",
		"aaa",
		"@@@",
	} {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixOfferToBuy, []byte(input)...), OfferToBuyKey(input))
			require.Equal(t, append(KeyPrefixRollAppIdToAlias, []byte(input)...), RollAppIdToAliasKey(input))
			require.Equal(t, append(KeyPrefixRvlAliasToRollAppId, []byte(input)...), AliasToRollAppIdRvlKey(input))
		})
	}
}
