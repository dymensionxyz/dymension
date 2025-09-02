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

					calculatedM := types.CalculateM(sdkmath.LegacyNewDec(raiseTarget), sdkmath.LegacyNewDec(allocation), tc.n, r)
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
						liquidityDenomDecimals,
					)
					t.Log("bootstrap funds", bootstrapFunds, "unsold value", unsoldValue, "leftover tokens", leftoverTokens, "curve price", curvePrice)
					// assert value of leftover tokens is the same as the bootstrap funds
					err := testutil.ApproxEqualRatio(bootstrapFunds, unsoldValue, 0.001) // 0.1%
					require.NoError(t, err)

					poolPrice := bootstrapFunds.ToLegacyDec().QuoInt(
						types.ScaleToBase(
							types.ScaleFromBase(leftoverTokens, 18),
							liquidityDenomDecimals,
						),
					)
					t.Log("curve price", curvePrice, "pool price", poolPrice)
					// assert price at eq is the same as the pool price
					err = testutil.ApproxEqualRatio(curvePrice, poolPrice, 0.001) // 0.1%
					require.NoError(t, err)

					totalValue := bootstrapFunds.Add(unsoldValue)
					targetRaiseScaled := types.ScaleToBase(sdkmath.LegacyNewDec(raiseTarget), liquidityDenomDecimals)
					t.Log("target raise", targetRaiseScaled, "total value", totalValue)
					// assert total value is same as expected
					err = testutil.ApproxEqualRatio(targetRaiseScaled, totalValue, 0.05) // 5% tolerance
					require.NoError(t, err)
				})
			})
		}
	}
}

func TestCalcLiquidityPoolTokens(t *testing.T) {
	// _ = flag.Set("rapid.checks", "1000") // can be enabled manually for more thorough testing

	// Define different curve types for generating realistic settled prices
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
					// Generate test parameters
					minAllocation := int64(1e4)  // 10K tokens
					maxAllocation := int64(1e8)  // 100M tokens
					minRaiseTarget := int64(1e4) // 10K DYM
					maxRaiseTarget := int64(1e7) // 10M DYM

					rFloat := rapid.Float64Range(0.1, 1).Draw(t, "bootstrap ratio")
					r := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", rFloat))

					// Generate a realistic settled token price using bonding curve logic
					allocation := testutil.LogarithmicRangeForRapid(t, minAllocation, maxAllocation)
					allocationScaled := sdkmath.NewInt(allocation).MulRaw(1e18)

					raiseTarget := testutil.LogarithmicRangeForRapid(t, minRaiseTarget, maxRaiseTarget)

					calculatedM := types.CalculateM(
						sdkmath.LegacyNewDec(raiseTarget),
						sdkmath.LegacyNewDec(allocation),
						tc.n,
						r,
					)
					if !calculatedM.IsPositive() {
						t.Skip("m is not positive")
					}

					curve := types.NewBondingCurve(calculatedM, tc.n, sdkmath.LegacyZeroDec(), 18, uint64(liquidityDenomDecimals))
					t.Logf("curve=%s, allocation=%d, target=%d, m=%s",
						tc.name, allocation, raiseTarget, calculatedM.String())

					eq := types.FindEquilibrium(curve, allocationScaled, r)
					require.True(t, eq.IsPositive())

					// test pool at eq point
					raisedLiquidityEq := curve.Cost(sdkmath.ZeroInt(), eq)
					poolTokens := raisedLiquidityEq.ToLegacyDec().Mul(r).TruncateInt()

					unsoldRATokensEq := allocationScaled.Sub(eq)
					spotPriceEq := curve.SpotPrice(eq)

					eqRATokens, eqLiquidity := types.CalcLiquidityPoolTokens(
						unsoldRATokensEq,
						poolTokens,
						spotPriceEq,
					)
					require.True(t, eqRATokens.LTE(unsoldRATokensEq), "eqRATokens:%s <= unsoldRATokensEq:%s", eqRATokens.String(), unsoldRATokensEq.String())
					require.True(t, eqLiquidity.LTE(poolTokens), "eqLiquidity:%s == poolTokens:%s", eqLiquidity.String(), poolTokens.String())
					err := testutil.ApproxEqualRatio(eqLiquidity, poolTokens, 0.001) // 0.1% tolerance
					require.NoError(t, err, "eqLiquidity should match poolTokens")

					// Verify the pool price relationship
					poolPrice := eqLiquidity.ToLegacyDec().Quo(eqRATokens.ToLegacyDec())
					err = testutil.ApproxEqualRatio(spotPriceEq, poolPrice, 0.001) // 0.1% tolerance
					require.NoError(t, err, "pool price should match settled token price")

					// Test a random point
					xRand := testutil.LogarithmicRangeForRapid(t, 1, types.ScaleFromBase(eq, 18).TruncateInt64())
					randomLiquidity := curve.Cost(sdkmath.ZeroInt(), sdkmath.NewInt(xRand))
					randomPoolTokens := randomLiquidity.ToLegacyDec().Mul(r).TruncateInt()

					unsoldRATokensRand := allocationScaled.Sub(sdkmath.NewInt(xRand))
					spotPriceRand := curve.SpotPrice(sdkmath.NewInt(xRand))

					randRATokens, randLiquidity := types.CalcLiquidityPoolTokens(
						unsoldRATokensRand,
						randomPoolTokens,
						spotPriceRand,
					)
					require.True(t, randRATokens.LTE(unsoldRATokensRand), "randRATokens:%s <= unsoldRATokensRand:%s", randRATokens.String(), unsoldRATokensRand.String())
					require.True(t, randLiquidity.LTE(randomPoolTokens), "randLiquidity:%s == randomPoolTokens:%s", randLiquidity.String(), randomPoolTokens.String())
					err = testutil.ApproxEqualRatio(randLiquidity, randomPoolTokens, 0.001) // 0.1% tolerance
					require.NoError(t, err, "randLiquidity should match randomPoolTokens")

					// Verify the pool price relationship at random point
					randomPoolPrice := randLiquidity.ToLegacyDec().Quo(randRATokens.ToLegacyDec())
					err = testutil.ApproxEqualRatio(spotPriceRand, randomPoolPrice, 0.001) // 0.1% tolerance
					require.NoError(t, err, "random pool price should match settled token price")
				})
			})
		}
	}
}

func TestFairLaunchEquilibrium(t *testing.T) {
	allocation := int64(1e9) // 1B RA tokens
	allocationScaled := sdkmath.NewInt(allocation).MulRaw(1e18)

	raiseTarget := int64(5 * 1e3) // 5K USD
	evaluation := raiseTarget * 2
	exponent := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("0.5"),
		sdkmath.LegacyMustNewDecFromStr("1.0"),
		sdkmath.LegacyMustNewDecFromStr("1.25"),
	}
	liquidityPart := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("1.0"),
	}

	for _, exponent := range exponent {
		for _, liquidityPart := range liquidityPart {
			t.Run(fmt.Sprintf("exponent=%s, liquidityPart=%s", exponent.String(), liquidityPart.String()), func(t *testing.T) {
				calculatedM := types.CalculateM(
					sdkmath.LegacyNewDec(evaluation),
					sdkmath.LegacyNewDec(allocation),
					exponent,
					liquidityPart,
				)
				require.True(t, calculatedM.IsPositive())

				curve := types.NewBondingCurve(calculatedM, exponent, sdkmath.LegacyZeroDec(), 18, 6)
				eq := types.FindEquilibrium(curve, allocationScaled, liquidityPart)
				require.True(t, eq.IsPositive())
				ratio := eq.ToLegacyDec().Quo(allocationScaled.ToLegacyDec())
				t.Logf("ratio=%s", ratio.String())

				// Due to Frontend restrictions, we want M to be > 10^-16
				require.True(t, calculatedM.GT(sdkmath.LegacyNewDecWithPrec(1, 16)), "calculatedM:%s", calculatedM.String())
			})
		}
	}
}
