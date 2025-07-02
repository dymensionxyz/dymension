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

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				minAllocation := int64(1e4) // 10K RA tokens
				maxAllocation := int64(1e9) // 1B RA tokens

				minRaiseTarget := int64(1e4) // 10K DYM
				maxRaiseTarget := int64(1e7) // 10M DYM

				rFloat := rapid.Float64Range(0.1, 1).Draw(t, "bootstrap ratio")
				r := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", rFloat))

				allocation := testutil.LogarithmicRangeForRapid(t, minAllocation, maxAllocation)
				allocationScaled := sdkmath.NewInt(allocation).MulRaw(1e18)

				raiseTarget := testutil.LogarithmicRangeForRapid(t, minRaiseTarget, maxRaiseTarget)
				raiseTargetDec := sdkmath.LegacyNewDec(raiseTarget)

				calcaulateM := types.CalculateM(raiseTargetDec, sdkmath.LegacyNewDec(allocation), tc.n, r)
				if !calcaulateM.IsPositive() {
					t.Skip("m is not positive", tc.name, "allocation", allocation, "targetRaise", raiseTarget)
				}
				t.Log("curve", tc.name, "allocation", allocation, "target", raiseTarget, "m", calcaulateM)

				curve := types.NewBondingCurve(calcaulateM, tc.n, sdkmath.LegacyZeroDec(), 18, 18)
				// assert eq is > 0
				eq := types.FindEquilibrium(curve, allocationScaled, r)
				require.True(t, eq.IsPositive())

				actualRaised := curve.Cost(sdkmath.ZeroInt(), eq)
				require.True(t, actualRaised.IsPositive())

				bootstrapFunds := actualRaised.ToLegacyDec().Mul(r).TruncateInt()
				require.True(t, bootstrapFunds.IsPositive())

				curvePrice := curve.SpotPrice(eq)
				leftoverTokens := allocationScaled.Sub(eq)

				// assert value of leftover tokens is the same as the bootstrap funds
				unsoldValue := curvePrice.MulInt(leftoverTokens).TruncateInt()
				err := testutil.ApproxEqualRatio(bootstrapFunds, unsoldValue, 0.001) // 0.1%
				require.NoError(t, err)

				// assert price at eq is the same as the pool price
				poolPrice := bootstrapFunds.ToLegacyDec().QuoInt(leftoverTokens)
				err = testutil.ApproxEqualRatio(curvePrice, poolPrice, 0.001) // 0.1%
				require.NoError(t, err)

				// assert total value is same as expected
				totalValue := bootstrapFunds.Add(unsoldValue)
				err = testutil.ApproxEqualRatio(raiseTargetDec.MulInt64(1e18).TruncateInt(), totalValue, 0.05) // 5% tolerance
				require.NoError(t, err)
			})
		})
	}
}
