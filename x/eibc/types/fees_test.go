package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestSolvePrice(t *testing.T) {

	target := math.NewInt(1000000000000000000)
	fee := math.NewInt(10000)
	bridgeFee, err := math.LegacyNewDecFromStr("0.01")
	require.NoError(t, err)

	amt := SolvePrice(target, fee, bridgeFee)

	eventualPrice, err := CalcPriceWithBridgingFee(amt, fee, bridgeFee)
	require.NoError(t, err)
	require.GreaterOrEqual(t, eventualPrice, target)
}
