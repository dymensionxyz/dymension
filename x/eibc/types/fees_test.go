package types

import (
	"flag"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestCalcTargetPriceAmt(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		target := math.NewInt(1000000000000000000)
		fee := math.NewInt(10000)
		bridgeFee, err := math.LegacyNewDecFromStr("0.01")
		require.NoError(t, err)
		testCalcTargetPriceAmt(t, target, fee, bridgeFee)
	})
	_ = flag.Set("rapid.checks", "200")
	rapid.Check(t, func(r *rapid.T) {
		target := math.NewInt(rapid.Int64Min(1).Draw(r, "target"))
		fee := math.NewInt(rapid.Int64Min(0).Draw(r, "fee"))
		bridgeFee := math.LegacyNewDecFromIntWithPrec(math.NewInt(rapid.Int64Range(0, 99).Draw(r, "bridgeFee")), 2)
		testCalcTargetPriceAmt(r, target, fee, bridgeFee)
	})
}

func testCalcTargetPriceAmt(t require.TestingT, target, fee math.Int, bridgeFee math.LegacyDec) {
	amt, err := CalcTargetPriceAmt(target, fee, bridgeFee)
	require.NoError(t, err)
	price, err := CalcPriceWithBridgingFee(amt, fee, bridgeFee)
	require.NoError(t, err)
	require.True(t, price.GTE(target), "price < target: %s < %s", price, target)
}
