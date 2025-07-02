package types_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestFindEquilibrium(t *testing.T) {
	// _ = flag.Set("rapid.checks", "1000") // can be enabled manually for more thorough testing

	// Define different curve types
	testcases := []struct {
		name string
		n    sdkmath.LegacyDec
	}{
		{"Square Root", sdkmath.LegacyMustNewDecFromStr("0.5")},
		{"Linear", sdkmath.LegacyOneDec()},
		{"Quadratic", sdkmath.LegacyMustNewDecFromStr("1.5")},
	}

	for _, liquidityDenomDecimals := range []int64{6, 18} {
		for _, tc := range testcases {
			tcName := fmt.Sprintf("%s-decimals=%d", tc.name, liquidityDenomDecimals)
			t.Run(tcName, func(t *testing.T) {
				rapid.Check(t, func(t *rapid.T) {
					minAllocation := int64(1e4) // 10K RA tokens
					// TODO: quadratic tests with HIGHER ALLOCATION starts to fail
					maxAllocation := int64(1e8) // 100M RA tokens

					minRaiseTarget := int64(1e4) // 10K DYM
					maxRaiseTarget := int64(1e7) // 10M DYM

					rFloat := rapid.Float64Range(0.1, 1).Draw(t, "bootstrap ratio")
					r := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", rFloat))

					allocation := testutil.LogarithmicRangeForRapid(t, minAllocation, maxAllocation)
					allocationScaled := sdkmath.NewInt(allocation).MulRaw(1e18)

					raiseTarget := testutil.LogarithmicRangeForRapid(t, minRaiseTarget, maxRaiseTarget)
					raiseTargetDec := sdkmath.LegacyNewDec(raiseTarget)

					calculatedM := types.CalculateM(raiseTargetDec, sdkmath.LegacyNewDec(allocation), tc.n, r)
					if !calculatedM.IsPositive() {
						t.Skip("m is not positive", tc.name, "allocation", allocation, "targetRaise", raiseTarget)
					}
					t.Log("curve", tc.name, "allocation", allocation, "target", raiseTarget, "m", calculatedM)

					curve := types.NewBondingCurve(calculatedM, tc.n, sdkmath.LegacyZeroDec(), 18, uint64(liquidityDenomDecimals))
					eq := types.FindEquilibrium(curve, allocationScaled, r)
					require.True(t, eq.IsPositive())

					actualRaised := curve.Cost(sdkmath.ZeroInt(), eq)
					require.True(t, actualRaised.IsPositive())

					bootstrapFunds := actualRaised.ToLegacyDec().Mul(r).TruncateInt()
					require.True(t, bootstrapFunds.IsPositive())

					curvePrice := curve.SpotPrice(eq)
					leftoverTokens := allocationScaled.Sub(eq)
					unsoldValue := types.ScaleToBase(
						curvePrice.Mul(types.ScaleFromBase(leftoverTokens, 18)),
						int64(liquidityDenomDecimals),
					)
					t.Log("bootstrap funds", bootstrapFunds, "unsold value", unsoldValue, "leftover tokens", leftoverTokens, "curve price", curvePrice)
					// assert value of leftover tokens is the same as the bootstrap funds
					err := testutil.ApproxEqualRatio(bootstrapFunds, unsoldValue, 0.001) // 0.1%
					require.NoError(t, err)

					poolPrice := bootstrapFunds.ToLegacyDec().QuoInt(
						types.ScaleToBase(
							types.ScaleFromBase(leftoverTokens, 18),
							int64(liquidityDenomDecimals),
						),
					)
					t.Log("curve price", curvePrice, "pool price", poolPrice)
					// assert price at eq is the same as the pool price
					err = testutil.ApproxEqualRatio(curvePrice, poolPrice, 0.001) // 0.1%
					require.NoError(t, err)

					totalValue := bootstrapFunds.Add(unsoldValue)
					targetRaiseScaled := types.ScaleToBase(raiseTargetDec, int64(liquidityDenomDecimals))
					t.Log("target raise", targetRaiseScaled, "total value", totalValue)
					// assert total value is same as expected
					err = testutil.ApproxEqualRatio(targetRaiseScaled, totalValue, 0.05) // 5% tolerance
					require.NoError(t, err)
				})
			})
		}
	}
}
