package types_test

import (
	"fmt"
	math "math"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// LogarithmicRange generates a random number in a logarithmic scale between min and max
func LogarithmicRange(t *rapid.T, min, max int64) int64 {
	if min <= 0 || max <= 0 || min >= max {
		panic("min and max must be positive and min must be less than max")
	}

	// Draw a random float in the range [0, 1]
	randomFloat := rapid.Float64Range(0, 1).Draw(t, "logRandom")

	// Apply logarithmic transformation
	logMin := math.Log(float64(min))
	logMax := math.Log(float64(max))

	// Scale the random float to the logarithmic range
	logValue := logMin + randomFloat*(logMax-logMin)

	return int64(math.Exp(logValue))
}

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

				allocation := LogarithmicRange(t, minAllocation, maxAllocation)
				allocationScaled := sdkmath.NewInt(allocation).MulRaw(1e18)

				raiseTarget := LogarithmicRange(t, minRaiseTarget, maxRaiseTarget)
				raiseTargetDec := sdkmath.LegacyNewDec(raiseTarget)

				t.Log("curve", tc.name, "allocation", allocation, "target", raiseTarget)

				calcaulateM := types.CalculateM(raiseTargetDec, sdkmath.LegacyNewDec(allocation), tc.n, r)
				if !calcaulateM.IsPositive() {
					t.Skip("m is not positive", tc.name, "allocation", allocation, "targetRaise", raiseTarget)
				}

				curve := types.NewBondingCurve(calcaulateM, tc.n, sdkmath.LegacyZeroDec(), 18, 18)
				// assert eq is > 0
				eq := types.FindEquilibrium(curve, allocationScaled, r)
				require.True(t, eq.IsPositive())

				actualRaised := curve.Cost(sdkmath.ZeroInt(), eq)
				require.True(t, actualRaised.IsPositive())

				bootstrapFunds := actualRaised.ToLegacyDec().Mul(r).TruncateInt()
				require.True(t, bootstrapFunds.IsPositive())

				// assert price at eq is the same as expected pool price
				curvePrice := curve.SpotPrice(eq)
				leftoverTokens := allocationScaled.Sub(eq)
				poolPrice := bootstrapFunds.ToLegacyDec().QuoInt(leftoverTokens)

				err := approxEqualRatio(curvePrice, poolPrice, 0.001) // 0.1%
				require.NoError(t, err)

				unsoldValue := curvePrice.MulInt(leftoverTokens).TruncateInt()
				err = approxEqualRatio(bootstrapFunds, unsoldValue, 0.001) // 0.1%
				require.NoError(t, err)

				// assert total value is same as expected
				totalValue := bootstrapFunds.Add(unsoldValue)
				err = approxEqualRatio(raiseTargetDec.MulInt64(1e18).TruncateInt(), totalValue, 0.01) // 5% tolerance
				require.NoError(t, err)
			})
		})
	}
}

// approxEqual checks if two values of different types are approximately equal
func approxEqual(expected, actual, tolerance interface{}) error {
	switch e := expected.(type) {
	case sdkmath.LegacyDec:
		a, ok := actual.(sdkmath.LegacyDec)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.LegacyDec")
		}
		tol, ok := tolerance.(sdkmath.LegacyDec)
		if !ok {
			return fmt.Errorf("tolerance is not a sdkmath.LegacyDec")
		}

		diff := e.Sub(a).Abs()
		if diff.GTE(tol) {
			return fmt.Errorf("expected %s, got %s, diff %s", e, a, diff)
		}

	case sdkmath.Int:
		a, ok := actual.(sdkmath.Int)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.Int")
		}
		tol, ok := tolerance.(sdkmath.Int)
		if !ok {
			return fmt.Errorf("tolerance is not a sdkmath.Int")
		}
		diff := e.Sub(a).Abs()
		if diff.GTE(tol) {
			return fmt.Errorf("expected %s, got %s, diff %s", e, a, diff)
		}
	default:
		return fmt.Errorf("unsupported type: %T", expected)
	}
	return nil
}

// approxEqualRatio checks if two values of different types are approximately equal
func approxEqualRatio(expected, actual interface{}, tolerance float64) error {
	switch e := expected.(type) {
	case sdkmath.LegacyDec:
		a, ok := actual.(sdkmath.LegacyDec)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.LegacyDec")
		}

		ratio := e.Quo(a).Abs().MustFloat64()
		if ratio < (1 - tolerance) {
			return fmt.Errorf("expected %s, got %s, diff %f", e, a, ratio)
		}

	case sdkmath.Int:
		a, ok := actual.(sdkmath.Int)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.Int")
		}

		ratio := e.ToLegacyDec().Quo(a.ToLegacyDec()).Abs().MustFloat64()
		if ratio < (1 - tolerance) {
			return fmt.Errorf("expected %s, got %s, diff %f", e, a, ratio)
		}

	default:
		return fmt.Errorf("unsupported type: %T", expected)
	}
	return nil
}
