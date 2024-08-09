// A copy of https://github.com/cosmos/cosmos-sdk/blob/v0.50.8/types/collections_test.go.

package collcompat

import (
	"testing"
	"time"

	"cosmossdk.io/collections/colltest"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestCollectionsCorrectness(t *testing.T) {
	t.Run("AccAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, AccAddressKey, sdk.AccAddress{0x0, 0x2, 0x3, 0x5})
	})

	t.Run("ValAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ValAddressKey, sdk.ValAddress{0x1, 0x3, 0x4})
	})

	t.Run("ConsAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ConsAddressKey, sdk.ConsAddress{0x32, 0x0, 0x0, 0x3})
	})

	t.Run("AddressIndexingKey", func(t *testing.T) {
		colltest.TestKeyCodec(t, LengthPrefixedAddressKey(AccAddressKey), sdk.AccAddress{0x2, 0x5, 0x8})
	})

	t.Run("Time", func(t *testing.T) {
		colltest.TestKeyCodec(t, TimeKey, time.Time{})
	})
}
