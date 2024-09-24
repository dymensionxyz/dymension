package types_test

import (
	fmt "fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

var (
	defaultTolerance = math.NewInt(1).MulRaw(1e9) // one millionth of a dym
)

// defaultApproxEqual
func approxEqual(t *testing.T, expected, actual math.Int) {
	diff := expected.Sub(actual).Abs()
	require.True(t, diff.LTE(defaultTolerance), fmt.Sprintf("expected %s, got %s, diff %s", expected, actual, diff))
}

// y=mx^n+c
// m >= 0, c > 0
func TestBondingCurve_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		m         float64
		n         float64
		c         float64
		expectErr bool
	}{
		{"Valid bonding curve", 2, 2.23, 3, false},
		{"Valid linear curve", 0.2, 0.88, 3.22, false},
		{"Valid const price curve", 0, 1, 3, false},
		{"Invalid C value", 2, 1, -1, true},
		{"Invalid M value", -2, 1, 3, true},
		{"Invalid N value", 2, -1, 3, true},
		{"Too high N value", 2, 11, 3, true},
		{"Precision check M", 2.222, 1, 3, true},
		{"Precision check N", 2, 1.2421, 3, true},
		{"Precision check C", 2, 1, 3.321312, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.m))
			n := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.n))
			c := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.c))

			bondingCurve := types.NewBondingCurve(m, n, c)
			err := bondingCurve.ValidateBasic()
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// simple linear curve
func TestBondingCurve_Linear(t *testing.T) {
	// y=x
	m := math.LegacyMustNewDecFromStr("1")
	n := math.LegacyMustNewDecFromStr("1")
	c := math.LegacyMustNewDecFromStr("0")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0).MulRaw(1e18)
	x2 := math.NewInt(10).MulRaw(1e18)
	x3 := math.NewInt(100).MulRaw(1e18)

	// Expected results
	spotPrice1 := math.NewInt(0).MulRaw(1e18)   // 1*0^1 + 0
	spotPrice2 := math.NewInt(10).MulRaw(1e18)  // 1*10^1 + 0
	spotPrice3 := math.NewInt(100).MulRaw(1e18) // 1*100^1 + 0

	// y = 1/2*x^2
	integral2 := math.NewInt(50).MulRaw(1e18)   // (1/2)*10^2
	integral3 := math.NewInt(5000).MulRaw(1e18) // (1/2)*100^2

	cost1to2 := integral2                      // 50 - 0
	cost2to3 := math.NewInt(4950).MulRaw(1e18) // 5000 - 50

	approxEqual(t, math.ZeroInt(), curve.Integral(x1))
	approxEqual(t, integral2, curve.Integral(x2))
	approxEqual(t, integral3, curve.Integral(x3))

	approxEqual(t, spotPrice1, curve.SpotPrice(x1))
	approxEqual(t, spotPrice2, curve.SpotPrice(x2))
	approxEqual(t, spotPrice3, curve.SpotPrice(x3))

	approxEqual(t, cost1to2, curve.Cost(x1, x2))
	approxEqual(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario 2: Quadratic Curve with Offset
func TestBondingCurve_Quadratic(t *testing.T) {
	// y=2x^2+10
	// integral of y = 2/3*x^3 + 10*x
	m := math.LegacyMustNewDecFromStr("2")
	n := math.LegacyMustNewDecFromStr("2")
	c := math.LegacyMustNewDecFromStr("10")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0).MulRaw(1e18)
	x2 := math.NewInt(5).MulRaw(1e18)
	x3 := math.NewInt(10).MulRaw(1e18)

	// Expected results
	spotPrice1 := math.NewInt(10).MulRaw(1e18)  // 2*0^2 + 10
	spotPrice2 := math.NewInt(60).MulRaw(1e18)  // 2*5^2 + 10
	spotPrice3 := math.NewInt(210).MulRaw(1e18) // 2*10^2 + 10

	integral1 := math.NewInt(0).MulRaw(1e18)                                                 // (2/3)*0^3 + 10*0
	integral2 := math.LegacyMustNewDecFromStr("133.3333333333").MulInt64(1e18).TruncateInt() // (2/3)*5^3 + 10*5                                                     // (2/3)*10^3 + 10*10
	integral3 := math.LegacyMustNewDecFromStr("766.6666666666").MulInt64(1e18).TruncateInt() // (2/3)*10^3 + 10*10

	cost1to2 := integral2                                                                   // (2/3)*5^3 + 10*5 - (2/3)*0^3 - 10*0
	cost2to3 := math.LegacyMustNewDecFromStr("633.3333333333").MulInt64(1e18).TruncateInt() // (2/3)*10^3 + 10*10 - (2/3)*5^3 - 10*5

	approxEqual(t, integral1, curve.Integral(x1))
	approxEqual(t, integral2, curve.Integral(x2))
	approxEqual(t, integral3, curve.Integral(x3))

	approxEqual(t, spotPrice1, curve.SpotPrice(x1))
	approxEqual(t, spotPrice2, curve.SpotPrice(x2))
	approxEqual(t, spotPrice3, curve.SpotPrice(x3))

	approxEqual(t, cost1to2, curve.Cost(x1, x2))
	approxEqual(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario 3: Cubic Curve with Large Numbers
func TestBondingCurve_Cubic(t *testing.T) {
	// y=3x^3+1000
	// integral of y = 3/4*x^4 + 1000*x
	m := math.LegacyMustNewDecFromStr("3")
	n := math.LegacyMustNewDecFromStr("3")
	c := math.LegacyMustNewDecFromStr("1000")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0).MulRaw(1e18)
	x2 := math.NewInt(100).MulRaw(1e18)
	x3 := math.NewInt(1000).MulRaw(1e18)

	// Expected results
	spotPrice1 := math.NewInt(1000).MulRaw(1e18)       // 3*0^3 + 1000
	spotPrice2 := math.NewInt(3001000).MulRaw(1e18)    // 3*100^3 + 1000
	spotPrice3 := math.NewInt(3000001000).MulRaw(1e18) // 3*1000^3 + 1000

	integral1 := math.NewInt(0).MulRaw(1e18)            // (3/4)*0^4 + 1000*0
	integral2 := math.NewInt(75100000).MulRaw(1e18)     // (3/4)*100^4 + 1000*100
	integral3 := math.NewInt(750001000000).MulRaw(1e18) // (3/4)*1000^4 + 1000*1000

	cost1to2 := math.NewInt(75100000).MulRaw(1e18)     // (3/4)*100^4 + 1000*100 - (3/4)*0^4 - 1000*0
	cost2to3 := math.NewInt(749925900000).MulRaw(1e18) // (3/4)*1000^4 + 1000*1000 - (3/4)*100^4 - 1000*100

	approxEqual(t, integral1, curve.Integral(x1))
	approxEqual(t, integral2, curve.Integral(x2))
	approxEqual(t, integral3, curve.Integral(x3))

	approxEqual(t, spotPrice1, curve.SpotPrice(x1))
	approxEqual(t, spotPrice2, curve.SpotPrice(x2))
	approxEqual(t, spotPrice3, curve.SpotPrice(x3))

	approxEqual(t, cost1to2, curve.Cost(x1, x2))
	approxEqual(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario: Square Root Curve
func TestBondingCurve_SquareRoot(t *testing.T) {
	// y = m*x^0.5 + c
	// integral of y = (2/3)*m*x^1.5 + c*x
	m := math.LegacyMustNewDecFromStr("2.24345436")
	n := math.LegacyMustNewDecFromStr("0.5")
	c := math.LegacyMustNewDecFromStr("10.5443534")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0).MulRaw(1e18)
	x2 := math.NewInt(100).MulRaw(1e18)
	x3 := math.NewInt(10000).MulRaw(1e18)

	// Expected results (rounded to nearest integer)
	spotPrice1 := math.LegacyMustNewDecFromStr("10.5443534").MulInt64(1e18).TruncateInt()  // 2.24345436*0^0.5 + 10.5443534 ≈ 11
	spotPrice2 := math.LegacyMustNewDecFromStr("32.978897").MulInt64(1e18).TruncateInt()   // 2.24345436*100^0.5 + 10.5443534 ≈ 33
	spotPrice3 := math.LegacyMustNewDecFromStr("234.8897894").MulInt64(1e18).TruncateInt() // 2.24345436*10000^0.5 + 10.5443534 ≈ 235

	integral1 := math.LegacyMustNewDecFromStr("0").MulInt64(1e18).TruncateInt()           // (2/3)*2.24345436*0^1.5 + 10.5443534*0 = 0
	integral2 := math.LegacyMustNewDecFromStr("2550.07158").MulInt64(1e18).TruncateInt()  // (2/3)*2.24345436*100^1.5 + 10.5443534*100 ≈ 2550
	integral3 := math.LegacyMustNewDecFromStr("1601079.774").MulInt64(1e18).TruncateInt() // (2/3)*2.24345436*10000^1.5 + 10.5443534*10000 ≈ 1598850

	cost1to2 := integral2                                                                  // integral2 - integral1
	cost2to3 := math.LegacyMustNewDecFromStr("1598529.70242").MulInt64(1e18).TruncateInt() // integral3 - integral2

	approxEqual(t, integral1, curve.Integral(x1))
	approxEqual(t, integral2, curve.Integral(x2))
	approxEqual(t, integral3, curve.Integral(x3))

	approxEqual(t, spotPrice1, curve.SpotPrice(x1))
	approxEqual(t, spotPrice2, curve.SpotPrice(x2))
	approxEqual(t, spotPrice3, curve.SpotPrice(x3))

	approxEqual(t, cost1to2, curve.Cost(x1, x2))
	approxEqual(t, cost2to3, curve.Cost(x2, x3))
}

/*

This function takes:
val: The total value to be raised (VAL)
z: The total number of tokens (Z)
k: The exponent (K)
It returns the calculated M value as a math.LegacyDec.


func CalculateM(val, z math.LegacyDec, k int64) math.LegacyDec {
    kPlusOne := math.LegacyNewDec(k + 1)
    zPowKPlusOne := z.Power(uint64(k + 1))
    return val.Mul(kPlusOne).Quo(zPowKPlusOne)
}

*/

/*
Real world scenario:
- A project wants to raise 100_000 DYM for 1_000_000 RA tokens
- N = 1
- C = 0.001 (1% of the average price)

Expected M value: 0.000000198
*/

func TestCalculateM(t *testing.T) {
	// Test case parameters
	val := math.LegacyNewDecFromInt(math.NewInt(100_000)) // 100,000 DYM to raise
	z := math.LegacyNewDecFromInt(math.NewInt(1_000_000)) // 1,000,000 RA tokens
	n := math.LegacyNewDec(1)                             // N = 1 (linear curve)
	c := math.LegacyNewDecWithPrec(1, 3)                  // C = 0.001 (1% of the average price)

	// Expected M calculation:

	expectedM := math.LegacyMustNewDecFromStr("0.000000198")

	// Calculate M
	m := types.CalculateM(val, z, n, c)
	require.Equal(t, expectedM, m)

	curve := types.NewBondingCurve(m, n, c)

	// Verify that the integral of the curve at Z equals VAL
	integral := curve.Integral(z.MulInt64(1e18).TruncateInt())
	approxEqual(t, val.MulInt64(1e18).TruncateInt(), integral)
}
