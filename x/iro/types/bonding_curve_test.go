package types_test

import (
	fmt "fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// approxEqualInt checks if two math.Ints are approximately equal
func approxEqualInt(t *testing.T, expected, actual math.Int) {
	defaultTolerance := math.NewInt(1e12) // one hundred-billionth of a dym
	diff := expected.Sub(actual).Abs()
	require.True(t, diff.LTE(defaultTolerance), fmt.Sprintf("expected %s, got %s, diff %s", expected, actual, diff))
}

// approxEqualDec checks if two math.Decs are approximately equal
func approxEqualDec(t *testing.T, expected, actual math.LegacyDec) {
	defaultTolerance := math.LegacyNewDecWithPrec(1, 12) // one hundred-billionth of a dym
	diff := expected.Sub(actual).Abs()
	require.True(t, diff.LTE(defaultTolerance), fmt.Sprintf("expected %s, got %s, diff %s", expected, actual, diff))
}

func TestBondingCurve_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		m         float64
		n         float64
		c         float64
		expectErr bool
	}{
		{"Valid bonding curve", 1, 1, 0, false},
		{"Valid linear curve", 0.000002, 1, 0.00022, false},
		{"Valid power curve N>1", 0.1234, 1.23, 0.002, false},
		{"Valid power curve N<1", 0.1234, 0.76, 0.002, false},
		{"Invalid C value", 2, 1, -1, true},
		{"Invalid M value", -2, 1, 3, true},
		{"Invalid N value", 2, -1, 3, true},
		{"Too high N value", 2, 11, 3, true},
		{"Precision check N", 2, 1.2421, 3, true},
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
	spotPrice1 := math.LegacyNewDec(0)   // 1*0^1 + 0
	spotPrice2 := math.LegacyNewDec(10)  // 1*10^1 + 0
	spotPrice3 := math.LegacyNewDec(100) // 1*100^1 + 0

	// y = 1/2*x^2
	integral2 := math.NewInt(50).MulRaw(1e18)   // (1/2)*10^2
	integral3 := math.NewInt(5000).MulRaw(1e18) // (1/2)*100^2

	cost1to2 := integral2                      // 50 - 0
	cost2to3 := math.NewInt(4950).MulRaw(1e18) // 5000 - 50

	approxEqualInt(t, math.ZeroInt(), curve.Cost(math.ZeroInt(), x1))
	approxEqualInt(t, math.ZeroInt(), curve.Cost(math.ZeroInt(), x1))
	approxEqualInt(t, integral2, curve.Cost(math.ZeroInt(), x2))
	approxEqualInt(t, integral2, curve.Cost(math.ZeroInt(), x2))
	approxEqualInt(t, integral3, curve.Cost(math.ZeroInt(), x3))

	approxEqualDec(t, spotPrice1, curve.SpotPrice(x1))
	approxEqualDec(t, spotPrice2, curve.SpotPrice(x2))
	approxEqualDec(t, spotPrice3, curve.SpotPrice(x3))

	approxEqualInt(t, cost1to2, curve.Cost(x1, x2))
	approxEqualInt(t, cost2to3, curve.Cost(x2, x3))
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
	spotPrice1 := math.LegacyNewDec(10)  // 2*0^2 + 10
	spotPrice2 := math.LegacyNewDec(60)  // 2*5^2 + 10
	spotPrice3 := math.LegacyNewDec(210) // 2*10^2 + 10

	integral1 := math.NewInt(0).MulRaw(1e18)                                                 // (2/3)*0^3 + 10*0
	integral2 := math.LegacyMustNewDecFromStr("133.3333333333").MulInt64(1e18).TruncateInt() // (2/3)*5^3 + 10*5                                                     // (2/3)*10^3 + 10*10
	integral3 := math.LegacyMustNewDecFromStr("766.6666666666").MulInt64(1e18).TruncateInt() // (2/3)*10^3 + 10*10

	cost1to2 := integral2                                                                   // (2/3)*5^3 + 10*5 - (2/3)*0^3 - 10*0
	cost2to3 := math.LegacyMustNewDecFromStr("633.3333333333").MulInt64(1e18).TruncateInt() // (2/3)*10^3 + 10*10 - (2/3)*5^3 - 10*5

	approxEqualInt(t, integral1, curve.Cost(math.ZeroInt(), x1))
	approxEqualInt(t, integral2, curve.Cost(math.ZeroInt(), x2))
	approxEqualInt(t, integral3, curve.Cost(math.ZeroInt(), x3))

	approxEqualDec(t, spotPrice1, curve.SpotPrice(x1))
	approxEqualDec(t, spotPrice2, curve.SpotPrice(x2))
	approxEqualDec(t, spotPrice3, curve.SpotPrice(x3))

	approxEqualInt(t, cost1to2, curve.Cost(x1, x2))
	approxEqualInt(t, cost2to3, curve.Cost(x2, x3))
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
	spotPrice1 := math.LegacyMustNewDecFromStr("10.5443534")  // 2.24345436*0^0.5 + 10.5443534 ≈ 11
	spotPrice2 := math.LegacyMustNewDecFromStr("32.978897")   // 2.24345436*100^0.5 + 10.5443534 ≈ 33
	spotPrice3 := math.LegacyMustNewDecFromStr("234.8897894") // 2.24345436*10000^0.5 + 10.5443534 ≈ 235

	integral1 := math.LegacyMustNewDecFromStr("0").MulInt64(1e18).TruncateInt()           // (2/3)*2.24345436*0^1.5 + 10.5443534*0 = 0
	integral2 := math.LegacyMustNewDecFromStr("2550.07158").MulInt64(1e18).TruncateInt()  // (2/3)*2.24345436*100^1.5 + 10.5443534*100 ≈ 2550
	integral3 := math.LegacyMustNewDecFromStr("1601079.774").MulInt64(1e18).TruncateInt() // (2/3)*2.24345436*10000^1.5 + 10.5443534*10000 ≈ 1598850

	cost1to2 := integral2                                                                  // integral2 - integral1
	cost2to3 := math.LegacyMustNewDecFromStr("1598529.70242").MulInt64(1e18).TruncateInt() // integral3 - integral2

	approxEqualInt(t, integral1, curve.Cost(math.ZeroInt(), x1))
	approxEqualInt(t, integral2, curve.Cost(math.ZeroInt(), x2))
	approxEqualInt(t, integral3, curve.Cost(math.ZeroInt(), x3))

	approxEqualDec(t, spotPrice1, curve.SpotPrice(x1))
	approxEqualDec(t, spotPrice2, curve.SpotPrice(x2))
	approxEqualDec(t, spotPrice3, curve.SpotPrice(x3))

	approxEqualInt(t, cost1to2, curve.Cost(x1, x2))
	approxEqualInt(t, cost2to3, curve.Cost(x2, x3))
}

// test very small x returns 0
func TestBondingCurve_SmallX(t *testing.T) {
	curve := types.DefaultBondingCurve()

	// less than 1 token is not enough
	require.True(t, curve.SpotPrice(math.NewInt(1_000_000)).IsZero())
	require.True(t, curve.Cost(math.ZeroInt(), math.NewInt(1_000_000)).IsZero())
	require.True(t, curve.Cost(math.ZeroInt(), math.NewInt(1).MulRaw(1e17)).IsZero())

	// even 1 token is enough
	require.False(t, curve.Cost(math.ZeroInt(), math.NewInt(1).MulRaw(1e18)).IsZero())
	require.False(t, curve.SpotPrice(math.NewInt(1).MulRaw(1e18)).IsZero())
}

// TestTokensForDYM tests the TokensForDYM function.
// This test suite performs the following steps for each test case:
// 1. Calculate the cost of buying a specified number of tokens using the classic Cost function.
// 2. Calculate the number of tokens that can be bought with the calculated cost.
// The goal is to ensure that both functions are inverses of each other.
func TestTokensForDYM(t *testing.T) {
	// Define multiple starting points (used as current sold amt)
	startingPoints := []string{"0", "0.5", "1", "100", "1000", "10000", "100000"}

	// Define multiple X token amounts to test (used as tokens to buy)
	xTokens := []string{"0.01", "0.1", "0.5", "1", "10", "1000", "10000", "100000", "1000000"}

	// Define different curve types
	curves := []struct {
		name  string
		curve types.BondingCurve
	}{
		{"Linear", types.DefaultBondingCurve()},
		{"Square Root", types.NewBondingCurve(
			math.LegacyMustNewDecFromStr("2.24345436"),
			math.LegacyMustNewDecFromStr("0.5"),
			math.LegacyMustNewDecFromStr("10.5443534"),
		)},
		{"Quadratic", types.NewBondingCurve(
			math.LegacyMustNewDecFromStr("2"),
			math.LegacyMustNewDecFromStr("1.5"),
			math.LegacyZeroDec(),
		)},
	}

	for _, curve := range curves {
		t.Run(curve.name, func(t *testing.T) {
			for _, start := range startingPoints {
				startingX := math.LegacyMustNewDecFromStr(start).MulInt64(1e18).TruncateInt()

				for _, xToken := range xTokens {
					x := math.LegacyMustNewDecFromStr(xToken).MulInt64(1e18).TruncateInt()
					expectedTokens := x

					cost := curve.curve.Cost(startingX, startingX.Add(x))
					if startingX.LT(math.NewInt(1e18)) {
						expectedTokens = math.ZeroInt()
					}

					tokens := curve.curve.TokensForExactDYM(startingX, cost)

					t.Run(fmt.Sprintf("Start=%s, X=%s", start, xToken), func(t *testing.T) {
						approxEqualInt(t, expectedTokens, tokens)
					})
				}
			}
		})
	}
}

/*
Real world scenario:
- A project wants to raise 100_000 DYM for 1_000_000 RA tokens
- N = 1
- C = 0.001 (1% of the average price)

Expected M value: 0.000000198
*/
func TestUseCaseA(t *testing.T) {
	// Test case parameters
	val := math.NewInt(100_000)          // 100,000 DYM to raise
	z := math.NewInt(1_000_000)          // 1,000,000 RA tokens
	n := math.LegacyNewDec(1)            // N = 1 (linear curve)
	c := math.LegacyNewDecWithPrec(1, 3) // C = 0.001 (1% of the average price)

	// Expected M calculation:
	expectedM := math.LegacyMustNewDecFromStr("0.000000198")

	// Calculate M
	m := types.CalculateM(math.LegacyNewDecFromInt(val), math.LegacyNewDecFromInt(z), n, c)
	require.Equal(t, expectedM, m)

	curve := types.NewBondingCurve(m, n, c)

	// Verify that the cost of the curve at Z equals VAL
	cost := curve.Cost(math.LegacyZeroDec().RoundInt(), z.MulRaw(1e18))
	approxEqualInt(t, val.MulRaw(1e18), cost)

	// verify that the cost early is lower than the cost later
	// test for buying 1000 RA tokens
	averagePrice := math.LegacyNewDecFromInt(val).QuoInt(z)
	costA := curve.Cost(math.ZeroInt(), math.NewInt(1000).MulRaw(1e18))
	costB := curve.Cost(math.NewInt(900_000).MulRaw(1e18), math.NewInt(901_000).MulRaw(1e18))
	t.Logf(
		"Average Price: %s DYM\nCost for 1k Tokens:\n  Yearly: %s DYM\n  90%%: %s DYM",
		averagePrice,
		costA.QuoRaw(1e18),
		costB.QuoRaw(1e18),
	)
	// Define a threshold for the cost difference (e.g., 5% of costA)
	threshold := costA.MulRaw(5).QuoRaw(100)

	// Assert that the cost difference is greater than the threshold
	costDifference := costB.Sub(costA)
	require.True(t, costDifference.GT(threshold),
		"Cost difference (%s) should be greater than threshold (%s)",
		costDifference, threshold)
}

func TestSpotPrice(t *testing.T) {
	t.Run("Constant Price Curve", func(t *testing.T) {
		// Test simplest case: y = 0.1 (constant price)
		m := math.LegacyZeroDec()
		n := math.LegacyZeroDec()
		c := math.LegacyMustNewDecFromStr("0.1")

		curve := types.NewBondingCurve(m, n, c)

		testCases := []struct {
			x    math.Int
			want math.LegacyDec
		}{
			{math.NewInt(0).MulRaw(1e18), c},    // expected price is 0.1
			{math.NewInt(1000).MulRaw(1e18), c}, // expected price is 0.1
		}

		for _, tc := range testCases {
			got := curve.SpotPrice(tc.x)
			assert.Equal(t, tc.want, got, "SpotPrice(%v) = %v, want %v", tc.x, got, tc.want)
		}
	})

	t.Run("Linear Price Curve", func(t *testing.T) {
		// Test linear case: y = 0.001x + 0.1
		m := math.LegacyMustNewDecFromStr("0.001")
		n := math.LegacyOneDec()
		c := math.LegacyMustNewDecFromStr("0.1")

		curve := types.NewBondingCurve(m, n, c)

		testCases := []struct {
			x    math.Int
			want math.LegacyDec
		}{
			{math.NewInt(0).MulRaw(1e18), math.LegacyMustNewDecFromStr("0.1")},
			{math.NewInt(1000).MulRaw(1e18), math.LegacyMustNewDecFromStr("1.1")},
		}

		for _, tc := range testCases {
			got := curve.SpotPrice(tc.x)
			assert.Equal(t, tc.want, got, "SpotPrice(%v) = %v, want %v", tc.x, got, tc.want)
		}
	})
}
