package types_test

import (
	fmt "fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/require"
)

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
		{"Valid bonding curve", 2, 2, 3, false},
		{"Valid linear curve", 2, 1, 3, false},
		{"Valid const price curve", 0, 1, 3, false},
		{"Invalid C value", 2, 1, 0, true},
		{"Invalid M value", -2, 1, 3, true},
		{"Invalid N value", 2, -1, 3, true},
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
	x1 := math.NewInt(0)
	x2 := math.NewInt(10)
	x3 := math.NewInt(100)

	// Expected results
	spotPrice1 := math.NewInt(0)   // 1*0^1 + 0
	spotPrice2 := math.NewInt(10)  // 1*10^1 + 0
	spotPrice3 := math.NewInt(100) // 1*100^1 + 0

	// y = 1/2*x^2
	integral2 := math.NewInt(50)   // (1/2)*10^2
	integral3 := math.NewInt(5000) // (1/2)*100^2

	cost1to2 := math.NewInt(50)   // 50 - 0
	cost2to3 := math.NewInt(4950) // 5000 - 50

	require.Equal(t, math.ZeroInt(), curve.Integral(x1))
	require.Equal(t, integral2, curve.Integral(x2))
	require.Equal(t, integral3, curve.Integral(x3))

	require.Equal(t, spotPrice1, curve.SpotPrice(x1))
	require.Equal(t, spotPrice2, curve.SpotPrice(x2))
	require.Equal(t, spotPrice3, curve.SpotPrice(x3))

	require.Equal(t, cost1to2, curve.Cost(x1, x2))
	require.Equal(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario 2: Quadratic Curve with Offset
func TestBondingCurve_Quadratic(t *testing.T) {
	// y=2x^2+10
	// integral of y = 2/3*x^3 + 10*x
	// m := math.NewInt(2)
	m := math.LegacyMustNewDecFromStr("2")
	n := math.LegacyMustNewDecFromStr("2")
	c := math.LegacyMustNewDecFromStr("10")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0)
	x2 := math.NewInt(5)
	x3 := math.NewInt(10)

	// Expected results
	spotPrice1 := math.NewInt(10)  // 2*0^2 + 10
	spotPrice2 := math.NewInt(60)  // 2*5^2 + 10
	spotPrice3 := math.NewInt(210) // 2*10^2 + 10

	integral1 := math.NewInt(0)   // (2/3)*0^3 + 10*0
	integral2 := math.NewInt(133) // (2/3)*5^3 + 10*5
	integral3 := math.NewInt(766) // (2/3)*10^3 + 10*10

	cost1to2 := math.NewInt(133) // (2/3)*5^3 + 10*5 - (2/3)*0^3 - 10*0
	cost2to3 := math.NewInt(633) // (2/3)*10^3 + 10*10 - (2/3)*5^3 - 10*5

	require.Equal(t, integral1, curve.Integral(x1))
	require.Equal(t, integral2, curve.Integral(x2))
	require.Equal(t, integral3, curve.Integral(x3))

	require.Equal(t, spotPrice1, curve.SpotPrice(x1))
	require.Equal(t, spotPrice2, curve.SpotPrice(x2))
	require.Equal(t, spotPrice3, curve.SpotPrice(x3))

	require.Equal(t, cost1to2, curve.Cost(x1, x2))
	require.Equal(t, cost2to3, curve.Cost(x2, x3))
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
	x1 := math.NewInt(0)
	x2 := math.NewInt(100)
	x3 := math.NewInt(1000)

	// Expected results
	spotPrice1 := math.NewInt(1000)       // 3*0^3 + 1000
	spotPrice2 := math.NewInt(3001000)    // 3*100^3 + 1000
	spotPrice3 := math.NewInt(3000001000) // 3*1000^3 + 1000

	integral1 := math.NewInt(0)            // (3/4)*0^4 + 1000*0
	integral2 := math.NewInt(75100000)     // (3/4)*100^4 + 1000*100
	integral3 := math.NewInt(750001000000) // (3/4)*1000^4 + 1000*1000

	cost1to2 := math.NewInt(75100000)     // (3/4)*100^4 + 1000*100 - (3/4)*0^4 - 1000*0
	cost2to3 := math.NewInt(749925900000) // (3/4)*1000^4 + 1000*1000 - (3/4)*100^4 - 1000*100

	require.Equal(t, integral1, curve.Integral(x1))
	require.Equal(t, integral2, curve.Integral(x2))
	require.Equal(t, integral3, curve.Integral(x3))

	require.Equal(t, spotPrice1, curve.SpotPrice(x1))
	require.Equal(t, spotPrice2, curve.SpotPrice(x2))
	require.Equal(t, spotPrice3, curve.SpotPrice(x3))

	require.Equal(t, cost1to2, curve.Cost(x1, x2))
	require.Equal(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario 4: High Exponent Curve
func TestBondingCurve_HighExponent(t *testing.T) {
	// y=x^5+100
	// integral of y = 1/6*x^6 + 100*x

	m := math.LegacyMustNewDecFromStr("1")
	n := math.LegacyMustNewDecFromStr("5")
	c := math.LegacyMustNewDecFromStr("100")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0)
	x2 := math.NewInt(2)
	x3 := math.NewInt(10)

	// Expected results
	spotPrice1 := math.NewInt(100)    // 1*0^5 + 100
	spotPrice2 := math.NewInt(132)    // 1*2^5 + 100
	spotPrice3 := math.NewInt(100100) // 1*10^5 + 100

	integral1 := math.NewInt(0)      // (1/6)*0^6 + 100*0
	integral2 := math.NewInt(210)    // (1/6)*2^6 + 100*2
	integral3 := math.NewInt(167666) // (1/6)*10^6 + 100*10

	cost1to2 := math.NewInt(210)    // 210 - 0
	cost2to3 := math.NewInt(167456) // 167666 - 210

	require.Equal(t, spotPrice1, curve.SpotPrice(x1))
	require.Equal(t, spotPrice2, curve.SpotPrice(x2))
	require.Equal(t, spotPrice3, curve.SpotPrice(x3))

	require.Equal(t, integral1, curve.Integral(x1))
	require.Equal(t, integral2, curve.Integral(x2))
	require.Equal(t, integral3, curve.Integral(x3))

	require.Equal(t, cost1to2, curve.Cost(x1, x2))
	require.Equal(t, cost2to3, curve.Cost(x2, x3))
}

// Scenario: Square Root Curve
func TestBondingCurve_SquareRoot(t *testing.T) {
	// y = m*x^0.5 + c
	// integral of y = (3/4)*m*x^1.5 + c*x
	m := math.LegacyMustNewDecFromStr("2.24345436")
	n := math.LegacyMustNewDecFromStr("0.5")
	c := math.LegacyMustNewDecFromStr("10.5443534")
	curve := types.NewBondingCurve(m, n, c)

	// Test values
	x1 := math.NewInt(0)
	x2 := math.NewInt(100)
	x3 := math.NewInt(10000)

	// Expected results (rounded to nearest integer)
	spotPrice1 := math.NewInt(11)  // 2.24345436*0^0.5 + 10.5443534 ≈ 11
	spotPrice2 := math.NewInt(33)  // 2.24345436*100^0.5 + 10.5443534 ≈ 33
	spotPrice3 := math.NewInt(235) // 2.24345436*10000^0.5 + 10.5443534 ≈ 235

	integral1 := math.NewInt(0)       // (2/3)*2.24345436*0^1.5 + 10.5443534*0 = 0
	integral2 := math.NewInt(2550)    // (2/3)*2.24345436*100^1.5 + 10.5443534*100 ≈ 2550
	integral3 := math.NewInt(1598850) // (2/3)*2.24345436*10000^1.5 + 10.5443534*10000 ≈ 1598850

	cost1to2 := math.NewInt(2550)    // integral2 - integral1
	cost2to3 := math.NewInt(1596300) // integral3 - integral2

	require.Equal(t, integral1, curve.Integral(x1))
	require.Equal(t, integral2, curve.Integral(x2))
	require.Equal(t, integral3, curve.Integral(x3))

	require.Equal(t, spotPrice1, curve.SpotPrice(x1))
	require.Equal(t, spotPrice2, curve.SpotPrice(x2))
	require.Equal(t, spotPrice3, curve.SpotPrice(x3))

	require.Equal(t, cost1to2, curve.Cost(x1, x2))
	require.Equal(t, cost2to3, curve.Cost(x2, x3))
}
