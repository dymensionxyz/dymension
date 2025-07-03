package math

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	"pgregory.net/rapid"
)

// LogarithmicRangeForRapid generates a random number in a logarithmic scale between min and max
func LogarithmicRangeForRapid(t *rapid.T, min, max int64) int64 {
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

// ApproxEqual checks if two values of different types are approximately equal
func ApproxEqual(expected, actual, allowed_diff interface{}) error {
	switch e := expected.(type) {
	case sdkmath.LegacyDec:
		a, ok := actual.(sdkmath.LegacyDec)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.LegacyDec")
		}
		tol, ok := allowed_diff.(sdkmath.LegacyDec)
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
		tol, ok := allowed_diff.(sdkmath.Int)
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

// ApproxEqualRatio checks if two values of different types are approximately equal
func ApproxEqualRatio(expected, actual interface{}, allowed_ratio_diff float64) error {
	switch e := expected.(type) {
	case sdkmath.LegacyDec:
		a, ok := actual.(sdkmath.LegacyDec)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.LegacyDec")
		}

		if a.IsZero() {
			if e.IsZero() {
				return nil // both zero, considered equal
			}
			return fmt.Errorf("expected %s, got zero", e)
		}

		// Convert allowed_ratio_diff to LegacyDec for precise comparison
		tolerance := sdkmath.LegacyNewDecWithPrec(int64(allowed_ratio_diff*1e18), 18)

		// Calculate |expected/actual - 1|
		ratio := e.Quo(a)
		one := sdkmath.LegacyOneDec()
		diff := ratio.Sub(one).Abs()

		if diff.GT(tolerance) {
			return fmt.Errorf("expected %s, got %s, ratio diff %s", e, a, diff)
		}

	case sdkmath.Int:
		a, ok := actual.(sdkmath.Int)
		if !ok {
			return fmt.Errorf("actual is not a sdkmath.Int")
		}

		if a.IsZero() {
			if e.IsZero() {
				return nil // both zero, considered equal
			}
			return fmt.Errorf("expected %s, got zero", e)
		}

		// Convert to LegacyDec for precise division
		expectedDec := e.ToLegacyDec()
		actualDec := a.ToLegacyDec()
		tolerance := sdkmath.LegacyNewDecWithPrec(int64(allowed_ratio_diff*1e18), 18)

		// Calculate |expected/actual - 1|
		ratio := expectedDec.Quo(actualDec)
		one := sdkmath.LegacyOneDec()
		diff := ratio.Sub(one).Abs()

		if diff.GT(tolerance) {
			return fmt.Errorf("expected %s, got %s, ratio diff %s", e, a, diff)
		}

	default:
		return fmt.Errorf("unsupported type: %T", expected)
	}
	return nil
}
