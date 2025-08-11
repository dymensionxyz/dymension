package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/osmosis-labs/osmosis/osmomath"
)

/*
A bonding curve is a mathematical function that defines the relationship between the price and supply of a token.
The general form of a bonding curve is typically expressed as P = M * x^N + C, where:
- P is the price
- X is the supply
- M is a multiplier that affects the curve's steepness
- N is the exponent that determines the curve's shape
- C is a constant that sets the initial price

The implications of these parameters are significant:
M (multiplier) controls the overall steepness of the curve. A higher M value results in a steeper price increase as supply grows, potentially leading to more rapid value appreciation but also higher volatility.

N (exponent) shapes the curve's trajectory. When N > 1, the curve becomes convex, accelerating price growth at higher supply levels, which can create strong incentives for early adoption. When 0 < N < 1, the curve is concave, slowing price growth as supply increases, which can promote more stable long-term growth.

C (constant) sets the starting price when supply is zero, effectively establishing a price floor and influencing the token's initial accessibility.
*/

const (
	MaxNValue     = 2 // Maximum allowed value for the N parameter
	MaxNPrecision = 3 // Maximum allowed decimal precision for the N parameter

	maxIterations    = 100 // maximum number of iterations for Newton-Raphson approximation
	epsilonPrecision = 12  // approximation precision decimal places (10^12)
)

/*
The bonding curve implementation based on decimal representation of the X (rollapp's tokens) and Y (liquidity) values.
we use scaling functions to convert between the decimal scale and the base denomination.
*/

func NewBondingCurve(m, n, c math.LegacyDec, rollappDenomDecimals, liquidityDenomDecimals uint64) BondingCurve {
	return BondingCurve{
		M:                      m,
		N:                      n,
		C:                      c,
		RollappDenomDecimals:   rollappDenomDecimals,
		LiquidityDenomDecimals: liquidityDenomDecimals,
	}
}

func DefaultBondingCurve() BondingCurve {
	// linear bonding curve as default
	return BondingCurve{
		M:                      math.LegacyMustNewDecFromStr("0.005"),
		N:                      math.LegacyOneDec(),
		C:                      math.LegacyZeroDec(),
		RollappDenomDecimals:   18,
		LiquidityDenomDecimals: 18,
	}
}

func (lbc BondingCurve) SupplyDecimals() int64 {
	return int64(lbc.RollappDenomDecimals) // nolint: gosec
}

func (lbc BondingCurve) LiquidityDecimals() int64 {
	return int64(lbc.LiquidityDenomDecimals) // nolint: gosec
}

// ValidateBasic checks if the bonding curve is valid
func (lbc BondingCurve) ValidateBasic() error {
	if lbc.M.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "m: %d", lbc.M)
	}
	if !lbc.N.IsPositive() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "n: %d", lbc.N)
	}
	if lbc.N.GT(math.LegacyNewDec(MaxNValue)) {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "n exceeds maximum value of %d: %s", MaxNValue, lbc.N)
	}

	if lbc.C.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "c: %s", lbc.C.String())
	}

	// positive C is supported only for fixed price for now (due to equilibrium calculation)
	if !lbc.C.IsZero() && !lbc.M.IsZero() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "m: %s, c: %s", lbc.M.String(), lbc.C.String())
	}

	if !checkPrecision(lbc.N) {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "N must have at most %d decimal places", MaxNPrecision)
	}

	if lbc.RollappDenomDecimals == 0 || lbc.LiquidityDenomDecimals == 0 {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "rollapp_basedenom_decimals: %d, liquidity_denom_decimals: %d", lbc.RollappDenomDecimals, lbc.LiquidityDenomDecimals)
	}

	return nil
}

/* ---------------------------------- APIs ---------------------------------- */
// APIs provide a way to interact with the bonding curve
// The inputs are provided in the base denomination

// SpotPrice returns the spot price at x
// - x: the current supply, in the base denomination
// - returns: the spot price at x, as price per token (e.g 0.1 DYM per token)
func (lbc BondingCurve) SpotPrice(x math.Int) math.LegacyDec {
	return lbc.spotPriceInternal(ScaleFromBase(x, lbc.SupplyDecimals()))
}

/*
The cost to purchase tokens from supply S1 to S2 is given by the definite integral of this function from S1 to S2. Mathematically, this is expressed as:
Cost = ∫(S1 to S2) (M * S^N + C) dS
Solving this integral yields:
Cost = [M / (N + 1) * S^(N + 1) + C * S](S1 to S2)

// - x: the current supply, in the base denomination
// - x1: the new supply, in the base denomination
// - returns: the cost to purchase tokens from x to x1, in adym
*/
func (lbc BondingCurve) Cost(x, x1 math.Int) math.Int {
	cost := lbc.integral(ScaleFromBase(x1, lbc.SupplyDecimals())).
		Sub(lbc.integral(ScaleFromBase(x, lbc.SupplyDecimals())))
	return ScaleToBase(cost, lbc.LiquidityDecimals())
}

// Calculate the number of tokens that can be bought for a given amount of liquidity
// As the integral of the bonding curve function is not invertible, we use the Newton-Raphson method to approximate the solution
// - currX: the current supply, in the base denomination
// - spendAmt: the amount of liquidity tokens to spend, in base denomination
// - returns: the number of tokens that can be bought with spendAmt, in the base denomination
func (lbc BondingCurve) TokensForExactInAmount(currX, spendAmt math.Int) (math.Int, error) {
	startingX := ScaleFromBase(currX, lbc.SupplyDecimals())
	spendTokens := ScaleFromBase(spendAmt, lbc.LiquidityDecimals())

	// If the current supply is less than 1, return 0
	if startingX.LT(math.LegacyOneDec()) {
		return math.ZeroInt(), errors.New("current supply is less than 1")
	}

	// If the spend amount is not positive, return 0
	if !spendAmt.IsPositive() {
		return math.ZeroInt(), errors.New("spend amount is not positive")
	}

	tokens, _, err := lbc.TokensApproximation(startingX, spendTokens)
	if err != nil {
		return math.ZeroInt(), err
	}

	return ScaleToBase(tokens, lbc.SupplyDecimals()), nil
}

/* --------------------------- internal functions --------------------------- */
// Calculate the number of tokens that can be bought with a given amount of liquidity tokens
// inputs validated and scaled by caller
func (lbc BondingCurve) TokensApproximation(startingX, spendTokens math.LegacyDec) (math.LegacyDec, int, error) {
	// Define the function we're trying to solve: f(x) = Integral(startingX + x) - Integral(startingX) - spendAmt
	f := func(x math.LegacyDec) math.LegacyDec {
		newX := startingX.Add(x)
		return lbc.integral(newX).Sub(lbc.integral(startingX)).Sub(spendTokens)
	}

	// Define the derivative of the function: f'(x) = SpotPrice(startingX + x)
	fPrime := func(x math.LegacyDec) math.LegacyDec {
		newX := startingX.Add(x)
		return lbc.spotPriceInternal(newX)
	}

	// Initial guess for the solution to the bonding curve equation
	x := spendTokens.Quo(lbc.spotPriceInternal(startingX))

	// Newton-Raphson iteration
	epsilonDec := math.LegacyNewDecWithPrec(1, epsilonPrecision)
	for i := 0; i < maxIterations; i++ {
		fx := f(x) // diff between spendTokens and the actual cost to get to x
		// If the function converges, return the result
		if !fx.IsPositive() && fx.Abs().LT(epsilonDec) {
			return x, i, nil
		}
		prevX := x
		fPrimex := fPrime(x) // price for new X

		// defensive check to avoid division by zero
		// not supposed to happen, as spotPriceInternal should never return 0
		if fPrimex.IsZero() {
			return math.LegacyDec{}, i, errors.New("division by zero")
		}
		x = x.Sub(fx.Quo(fPrimex))

		// If the change in x is less than epsilon * x, return the result
		if x.Sub(prevX).Abs().LT(epsilonDec.Mul(x.Abs())) {
			return x, i, nil
		}

		// we can't allow newX to be less than 1
		if startingX.Add(x).LT(math.LegacyOneDec()) {
			x = math.LegacyOneDec()
		}
	}
	return math.LegacyDec{}, maxIterations, errors.New("solution did not converge")
}

// spotPriceInternal returns the spot price at x
func (lbc BondingCurve) spotPriceInternal(x math.LegacyDec) math.LegacyDec {
	xDec := osmomath.BigDecFromSDKDec(x)
	nDec := osmomath.BigDecFromSDKDec(lbc.N)
	mDec := osmomath.BigDecFromSDKDec(lbc.M)

	var xPowN osmomath.BigDec
	if xDec.LT(osmomath.OneDec()) {
		xPowN = osmomath.ZeroDec()
	} else {
		xPowN = xDec.Power(nDec) // Calculate x^N
	}
	price := mDec.Mul(xPowN).SDKDec().Add(lbc.C) // M * x^N + C
	return price
}

// The integral of y = M * x^N + C is:
//
//	Cost = (M / (N + 1)) * x^(N + 1) + C * x.
func (lbc BondingCurve) integral(x math.LegacyDec) math.LegacyDec {
	xDec := osmomath.BigDecFromSDKDec(x)
	mDec := osmomath.BigDecFromSDKDec(lbc.M)
	cDec := osmomath.BigDecFromSDKDec(lbc.C)

	nPlusOne := osmomath.BigDecFromSDKDec(lbc.N.Add(math.LegacyNewDec(1)))

	var xPowNplusOne osmomath.BigDec
	if xDec.LT(osmomath.OneDec()) {
		xPowNplusOne = osmomath.ZeroDec()
	} else {
		xPowNplusOne = xDec.Power(nPlusOne) // Calculate x^(N + 1)
	}

	mDivNPlusOne := mDec.QuoMut(nPlusOne) // Calculate m / (N + 1)
	cx := cDec.Mul(xDec)                  // Calculate C * x

	// Calculate the integral
	integral := xPowNplusOne.Mul(mDivNPlusOne).Add(cx).SDKDec()
	return integral
}

// CalculateM computes the M parameter for a bonding curve
// It's actually not used in the codebase, but it's here for reference and for testing purposes
// val: total value to be raised in display denom (e.g DYM, not adym)
// t: total number of tokens (rollapp's tokens in decimal representation, not base denomination)
// n: curve exponent
//
// we use the eq point (eq = ((N+1) * T) / (N+2)) to calculate M
// The total value at the eq, which consists of raised liquidity and unsold tokens, should equal val
// solving the equation for M gives:
// M = (VAL * (N+1) * (R + N + 1)^(N+1)) / (2 * R * ((N+1) * T)^(N+1))

func CalculateM(val, t, n, r math.LegacyDec) math.LegacyDec {
	valBig := osmomath.BigDecFromSDKDec(val)
	tBig := osmomath.BigDecFromSDKDec(t)
	nBig := osmomath.BigDecFromSDKDec(n)
	rBig := osmomath.BigDecFromSDKDec(r)

	// Calculate N + 1
	nPlusOne := nBig.Add(osmomath.OneDec())
	nPlusTwo := nPlusOne.Add(rBig)

	// we solve the equation logarithmically, as T^(N+1) can cause truncations

	// log(nominator) = log(val) + log(N + 1) + (N+1)log(N + 1 + R)
	lognum := valBig.LogBase2().Add(nPlusOne.LogBase2()).Add(nPlusOne.Mul(nPlusTwo.LogBase2()))

	// log(denominator) = (N+1)*log(T*(N+1)) + log(2) + log(R)
	logdenom := (nPlusOne.Mul((tBig.Mul(nPlusOne)).LogBase2())).Add(osmomath.OneDec()).Add(rBig.LogBase2())

	logm := lognum.Sub(logdenom)
	m := osmomath.Exp2(logm.Abs())

	if logm.IsNegative() {
		m = osmomath.OneDec().Quo(m)
	}

	// Convert back to math.LegacyDec and return
	return m.SDKDec()
}

/* ---------------------------- helper functions ---------------------------- */
// Scales x from it's base denomination to a decimal representation (e.g 1500000000000000 to 1.5)
// This is used to scale X before passing it to the bonding curve functions
func ScaleFromBase(x math.Int, precision int64) math.LegacyDec {
	return math.LegacyNewDecFromIntWithPrec(x, precision)
}

// Scales x from the decimal scale to it's base denomination (e.g 1.5 to 1500000000000000)
func ScaleToBase(x math.LegacyDec, precision int64) math.Int {
	scaleFactor := math.NewIntWithDecimal(1, int(precision))
	return x.MulInt(scaleFactor).TruncateInt()
}

// checkPrecision checks if a math.LegacyDec has at most MaxPrecision decimal places
func checkPrecision(d math.LegacyDec) bool {
	// Multiply by 10^MaxPrecision and check if it's an integer
	multiplied := d.Mul(math.LegacyNewDec(10).Power(uint64(MaxNPrecision)))
	return multiplied.IsInteger()
}

// String returns a human readable string representation of the bonding curve
func (lbc BondingCurve) Stringify() string {
	return fmt.Sprintf("M=%s N=%s C=%s",
		lbc.M.String(),
		lbc.N.String(),
		lbc.C.String(),
	)
}
