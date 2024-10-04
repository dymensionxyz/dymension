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
		require.Equal(t, []byte{0x05, partialStoreAssetTypeDymName}, KeyPrefixDymNameSellOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x05, partialStoreAssetTypeAlias}, KeyPrefixAliasSellOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x08}, KeyPrefixBuyOrder, "do not change it, will break the app")
		require.Equal(t, []byte{0x09}, KeyPrefixRvlBuyerToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0A, partialStoreAssetTypeDymName}, KeyPrefixRvlDymNameToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0A, partialStoreAssetTypeAlias}, KeyPrefixRvlAliasToBuyOrderIds, "do not change it, will break the app")
		require.Equal(t, []byte{0x0B}, KeyPrefixRollAppIdToAliases, "do not change it, will break the app")
		require.Equal(t, []byte{0x0C}, KeyPrefixRvlAliasToRollAppId, "do not change it, will break the app")
	})

	t.Run("ensure keys are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, []byte{0x07}, KeyCountBuyOrders, "do not change it, will break the app")
	})

	t.Run("ensure partitioned keys are not mistakenly modified", func(t *testing.T) {
		require.Equal(t, byte(0x00), byte(partialStoreAssetTypeDymName), "do not change it, will break the app")
		require.Equal(t, byte(0x01), byte(partialStoreAssetTypeAlias), "do not change it, will break the app")
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeys(t *testing.T) {
	for _, dymName := range []string{"a", "b", "my-name"} {
		t.Run(dymName, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixDymName, []byte(dymName)...), DymNameKey(dymName))
			require.Equal(t, append(KeyPrefixDymNameSellOrder, []byte(dymName)...), SellOrderKey(dymName, TypeName))
			require.Equal(t, append(KeyPrefixRvlDymNameToBuyOrderIds, []byte(dymName)...), DymNameToBuyOrderIdsRvlKey(dymName))
		})
	}

	for _, alias := range []string{"a", "b", "alias"} {
		t.Run(alias, func(t *testing.T) {
			require.Equal(t, append(KeyPrefixAliasSellOrder, []byte(alias)...), SellOrderKey(alias, TypeAlias))
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

	t.Run("should panics of getting Sell-Order related keys if asset type is invalid", func(t *testing.T) {
		require.Panics(t, func() { _ = SellOrderKey("asset", AssetType_AT_UNKNOWN) })
	})
}
