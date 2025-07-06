package types_test

import (
	"errors"
	fmt "fmt"
	"sort"
	"testing"

	"cosmossdk.io/math"
	testutil "github.com/dymensionxyz/dymension/v3/testutil/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

var (
	defaultToleranceInt = math.NewIntWithDecimal(1, 12)    // one millionth of a dym
	defaultToleranceDec = math.LegacyNewDecWithPrec(1, 12) // one millionth of a dym
)

// approxEqualInt checks if two math.Ints are approximately equal
func approxEqualInt(t *testing.T, expected, actual math.Int) {
	err := testutil.ApproxEqual(expected, actual, defaultToleranceInt)
	assert.NoError(t, err)
}

// approxEqualDec checks if two math.Decs are approximately equal
func approxEqualDec(t *testing.T, expected, actual math.LegacyDec) {
	err := testutil.ApproxEqual(expected, actual, defaultToleranceDec)
	assert.NoError(t, err)
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
		{"Valid fixed curve", 0, 1, 0.15, false},
		{"Valid linear curve", 0.000002, 1, 0, false},
		{"Valid power curve N>1", 0.1234, 1.23, 0.00, false},
		{"Valid power curve N<1", 0.1234, 0.76, 0.00, false},
		{"Invalid C value", 2, 1, -1, true},
		{"Invalid M value", -2, 1, 0, true},
		{"Invalid N value", 2, -1, 0, true},
		{"Too high N value", 2, 11, 0, true},
		{"Precision check N", 2, 1.2421, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.m))
			n := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.n))
			c := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", tt.c))

			bondingCurve := types.NewBondingCurve(m, n, c, 18, 18)
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
	curve := types.NewBondingCurve(m, n, c, 18, 18)

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
	curve := types.NewBondingCurve(m, n, c, 18, 18)

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
	curve := types.NewBondingCurve(m, n, c, 18, 18)

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
	startingPoints := []string{"1", "100", "1000", "10000", "100000"}

	// Define multiple X token amounts to test (used as tokens to buy)
	xTokens := []string{"0.01", "0.1", "0.5", "1", "10", "1000", "10000", "100000", "1000000"}

	// Define different curve types
	testcases := []struct {
		name string
		n    math.LegacyDec
	}{
		{"Square Root", math.LegacyMustNewDecFromStr("0.5")},
		{"Linear", math.LegacyOneDec()},
		{"Quadratic", math.LegacyMustNewDecFromStr("1.5")},
	}

	for _, liquidityDenomDecimals := range []int64{6, 18} {
		for _, tc := range testcases {
			tcName := fmt.Sprintf("%s-decimals=%d", tc.name, liquidityDenomDecimals)
			t.Run(tcName, func(t *testing.T) {
				rapid.Check(t, func(t *rapid.T) {
					minAllocation := int64(1e4) // 10K RA tokens
					maxAllocation := int64(1e8) // 100M RA tokens

					minRaiseTarget := int64(1e4) // 10K DYM
					maxRaiseTarget := int64(1e7) // 10M DYM

					rFloat := rapid.Float64Range(0.1, 1).Draw(t, "bootstrap ratio")
					r := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", rFloat))

					allocation := testutil.LogarithmicRangeForRapid(t, minAllocation, maxAllocation)
					raiseTarget := testutil.LogarithmicRangeForRapid(t, minRaiseTarget, maxRaiseTarget)
					raiseTargetDec := math.LegacyNewDec(raiseTarget)

					calculatedM := types.CalculateM(raiseTargetDec, math.LegacyNewDec(allocation), tc.n, r)
					if !calculatedM.IsPositive() {
						t.Skip("m is not positive", tc.name, "allocation", allocation, "targetRaise", raiseTarget)
					}

					t.Logf("curve=%s, allocation=%d, target=%d, m=%s",
						tc.name, allocation, raiseTarget, calculatedM.String())

					curve := types.NewBondingCurve(calculatedM, tc.n, math.LegacyZeroDec(), 18, uint64(liquidityDenomDecimals))

					// Test with predefined starting points and token amounts
					for _, start := range startingPoints {
						startingX := math.LegacyMustNewDecFromStr(start).MulInt64(1e18).TruncateInt()

						check := false
						for _, xToken := range xTokens {
							x := math.LegacyMustNewDecFromStr(xToken).MulInt64(1e18).TruncateInt()
							cost := curve.Cost(startingX, startingX.Add(x))
							// skip if cost is less than 0.00005USDC or 0.000000000005DYM
							if cost.LT(math.NewInt(50).Mul(math.NewIntWithDecimal(1, int(liquidityDenomDecimals)-6))) {
								t.Logf("cost is less than 50, skipping startingX=%s, xToken=%s, cost=%s", start, xToken, cost)
								continue
							}

							tokens, err := curve.TokensForExactInAmount(startingX, cost)
							require.NoError(t, err)

							errRatio := testutil.ApproxEqualRatio(x, tokens, 0.05)                   // 5% tolerance
							errInt := testutil.ApproxEqual(x, tokens, math.NewIntWithDecimal(5, 15)) // 0.005 RA token
							if errRatio != nil && errInt != nil {
								assert.NoError(t, errors.Join(errRatio, errInt),
									fmt.Sprintf("startingX=%s, xToken=%s, cost=%s, tokens=%s", start, xToken, cost, types.ScaleFromBase(tokens, 18)))
							}
							check = true
						}
						require.True(t, check, "no check was done for startingX=%s", start)
					}
				})
			})
		}
	}
}

// benchmark the iteration count for the TokensForDYM function
func TestTokensForDYMApproximation(t *testing.T) {
	// _ = flag.Set("rapid.checks", "10000") // can be enabled manually for more thorough testing

	// Define different curve types
	curves := []struct {
		name  string
		curve types.BondingCurve
	}{
		{"Linear", types.DefaultBondingCurve()},
		{"Square Root", types.NewBondingCurve(
			math.LegacyMustNewDecFromStr("2.24345436"),
			math.LegacyMustNewDecFromStr("0.5"),
			math.LegacyMustNewDecFromStr("0.0"),
			18,
			18,
		)},
		{"Quadratic", types.NewBondingCurve(
			math.LegacyMustNewDecFromStr("2"),
			math.LegacyMustNewDecFromStr("1.5"),
			math.LegacyMustNewDecFromStr("0.0"),
			18,
			18,
		)},
	}

	for _, curve := range curves {
		var iterations []int

		t.Run(curve.name, func(t *testing.T) {
			rapid.Check(t, func(t *rapid.T) {
				startingXRapid := rapid.Int64Range(1, 1e6).Draw(t, "startingX")
				xRapid := rapid.Float64Range(0.01, 1e6).Draw(t, "x")

				startingX := math.LegacyNewDec(startingXRapid).MulInt64(1e18).TruncateInt()
				x := math.LegacyMustNewDecFromStr(fmt.Sprintf("%f", xRapid)).MulInt64(1e18).TruncateInt()

				cost := curve.curve.Cost(startingX, startingX.Add(x))

				startingXScaled := types.ScaleFromBase(startingX, curve.curve.SupplyDecimals())
				spendTokensScaled := types.ScaleFromBase(cost, 18)
				_, iteration, err := curve.curve.TokensApproximation(startingXScaled, spendTokensScaled)
				require.NoError(t, err)

				if err != nil {
					t.Fatalf("Error in TokensApproximation: %v", err)
				}

				t.Logf("Start=%d, X=%f, Iteration=%d", startingX, x, iteration)
				iterations = append(iterations, iteration)
			})
		})

		// After all checks are done
		sort.Ints(iterations)
		min, max := iterations[0], iterations[len(iterations)-1]
		sum := 0
		for _, v := range iterations {
			sum += v
		}
		avg := float64(sum) / float64(len(iterations))

		t.Logf("Statistics for %s curve:", curve.name)
		t.Logf("  Min iterations: %d", min)
		t.Logf("  Max iterations: %d", max)
		t.Logf("  Average iterations: %.2f", avg)
	}
}

/*
Real world scenario:
- A project wants to raise 100_000 DYM for 1_000_000 RA tokens
- N = 1

Expected M value: 0.000000224999999999
*/
func TestUseCaseA(t *testing.T) {
	// Test case parameters
	val := math.NewInt(100_000) // 100,000 DYM to raise
	z := math.NewInt(1_000_000) // 1,000,000 RA tokens
	n := math.LegacyNewDec(1)   // N = 1 (linear curve)
	c := math.LegacyZeroDec()
	r := math.LegacyMustNewDecFromStr("0.9") // 10% goes to founder

	// Calculate M
	m := types.CalculateM(math.LegacyNewDecFromInt(val), math.LegacyNewDecFromInt(z), n, r)
	require.True(t, m.IsPositive())

	curve := types.NewBondingCurve(m, n, c, 18, 18)

	// find eq
	eq := types.FindEquilibrium(curve, z.MulRaw(1e18), r)

	// verify that the cost early is lower than the cost later
	// test for buying 10_000 RA tokens
	averagePrice := math.LegacyNewDecFromInt(val).QuoInt(z)
	costFirst := curve.Cost(math.ZeroInt(), math.NewInt(10_000).MulRaw(1e18))                   // first 10K tokens
	costEarly := curve.Cost(math.NewInt(10_000).MulRaw(1e18), math.NewInt(20_000).MulRaw(1e18)) // next 10K tokens
	costLast := curve.Cost(eq.Sub(math.NewInt(10_000).MulRaw(1e18)), eq)                        // last 10K tokens
	t.Logf(
		"Average Cost: %s DYM\nCost for 1k Tokens:\n, first: %s DYM\n  early: %s DYM\n  last: %s DYM",
		averagePrice.MulInt64(10_000),
		costFirst.ToLegacyDec().QuoInt64(1e18),
		costEarly.ToLegacyDec().QuoInt64(1e18),
		costLast.ToLegacyDec().QuoInt64(1e18),
	)

	// Define a threshold for the minimum price increase (10%)
	increaseFactor := math.LegacyMustNewDecFromStr("1.10")

	// Assert that each price is at least 10% higher than the previous
	require.True(t, costEarly.GTE(costFirst.ToLegacyDec().Mul(increaseFactor).TruncateInt()),
		"Early cost (%s) should be at least 10%% higher than first cost (%s)",
		costEarly, costFirst)

	increaseFactor = math.LegacyMustNewDecFromStr("1.50")
	require.True(t, costLast.GTE(costEarly.ToLegacyDec().Mul(increaseFactor).TruncateInt()),
		"Last cost (%s) should be at least 10%% higher than early cost (%s)",
		costLast, costEarly)

	// Validate that the TVL in the pool is correct
	bootstrapFunds := curve.Cost(math.ZeroInt(), eq).ToLegacyDec().Mul(r).TruncateInt()
	unsoldRATokens := z.MulRaw(1e18).Sub(eq)
	unsoldValue := curve.SpotPrice(eq).MulInt(unsoldRATokens).TruncateInt()

	// assert dym value in the pool is equal to unsold value
	err := testutil.ApproxEqualRatio(bootstrapFunds, unsoldValue, 0.001) // 0.1%
	require.NoError(t, err)

	// assert the TVL in eq point is as expected
	totalValue := bootstrapFunds.Add(unsoldValue)
	err = testutil.ApproxEqualRatio(val.MulRaw(1e18), totalValue, 0.001) // 0.1%
	require.NoError(t, err)
}

func TestUseCase_USDC(t *testing.T) {
	// Test case parameters
	val := math.NewInt(100_000) // 100,000 liquidity to raise
	z := math.NewInt(1_000_000) // 1,000,000 RA tokens
	n := math.LegacyNewDec(1)   // N = 1 (linear curve)
	c := math.LegacyZeroDec()
	r := math.LegacyMustNewDecFromStr("0.9") // 10% goes to founder

	// Calculate M
	m := types.CalculateM(math.LegacyNewDecFromInt(val), math.LegacyNewDecFromInt(z), n, r)
	require.True(t, m.IsPositive())

	curveUSDC := types.NewBondingCurve(m, n, c, 18, 6) // we set 18 for RA and 6 for USDC
	curveDYM := types.NewBondingCurve(m, n, c, 18, 18) // we set 18 for RA and 18 for DYM

	// find eq
	eq := types.FindEquilibrium(curveUSDC, z.MulRaw(1e18), r)

	// verify that the cost early is lower than the cost later
	// test for buying 10_000 RA tokens

	costUSDCFirst := curveUSDC.Cost(math.ZeroInt(), math.NewInt(10_000).MulRaw(1e18)) // first 10K tokens
	costDYMFirst := curveDYM.Cost(math.ZeroInt(), math.NewInt(10_000).MulRaw(1e18))

	require.Equal(t,
		types.ScaleFromBase(costUSDCFirst, 6).TruncateInt().String(),
		types.ScaleFromBase(costDYMFirst, 18).TruncateInt().String())

	costUSDCLast := curveUSDC.Cost(eq.Sub(math.NewInt(10_000).MulRaw(1e18)), eq) // last 10K tokens
	costDYMLast := curveDYM.Cost(eq.Sub(math.NewInt(10_000).MulRaw(1e18)), eq)

	require.Equal(t,
		types.ScaleFromBase(costUSDCLast, 6).TruncateInt().String(),
		types.ScaleFromBase(costDYMLast, 18).TruncateInt().String())
}

func TestSpotPrice(t *testing.T) {
	t.Run("Constant Price Curve", func(t *testing.T) {
		// Test simplest case: y = 0.1 (constant price)
		m := math.LegacyZeroDec()
		n := math.LegacyZeroDec()
		c := math.LegacyMustNewDecFromStr("0.1")

		curve := types.NewBondingCurve(m, n, c, 18, 18)

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

		curve := types.NewBondingCurve(m, n, c, 18, 18)

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
