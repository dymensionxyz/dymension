package types

import (
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
)

/*
The bonding curve implementation based on decimal representation of the X (rollapp's tokens) and Y (DYM) values.
we use scaling functions to convert between the decimal scale and the base denomination.
*/

// Scales x from it's base denomination to a decimal representation
// This is used so the bonding curve
func ScaleXFromBase(x math.Int, precision int64) math.LegacyDec {
	return math.LegacyNewDecFromIntWithPrec(x, precision)
}

// Scales y from the decimal scale to it's base denomination
func ScaleDYMToBase(y math.LegacyDec) math.Int {
	return y.MulInt(math.NewInt(1e18)).TruncateInt()
}

func (lbc BondingCurve) SupplyDecimals() int64 {
	// TODO: allow to be set on creation instead of using default
	rollappTokenDefaultDecimals := int64(18)
	return rollappTokenDefaultDecimals
}

func NewBondingCurve(m, n, c math.LegacyDec) BondingCurve {
	return BondingCurve{
		M: m,
		N: n,
		C: c,
	}
}

func DefaultBondingCurve() BondingCurve {
	// linear bonding curve as default
	return BondingCurve{
		M: math.LegacyMustNewDecFromStr("0.005"),
		N: math.LegacyOneDec(),
		C: math.LegacyZeroDec(),
	}
}

// validateBasic checks if the bonding curve is valid
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

	if !checkPrecision(lbc.N) {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "N must have at most %d decimal places", MaxNPrecision)
	}

	return nil
}

// checkPrecision checks if a math.LegacyDec has at most MaxPrecision decimal places
func checkPrecision(d math.LegacyDec) bool {
	// Multiply by 10^MaxPrecision and check if it's an integer
	multiplied := d.Mul(math.LegacyNewDec(10).Power(uint64(MaxNPrecision)))
	return multiplied.IsInteger()
}

// SpotPrice returns the spot price at x
func (lbc BondingCurve) SpotPrice(x math.Int) math.LegacyDec {
	// we use osmomath as it support Power function
	xDec := osmomath.BigDecFromSDKDec(ScaleXFromBase(x, lbc.SupplyDecimals()))
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

/*
The cost to purchase tokens from supply S1 to S2 is given by the definite integral of this function from S1 to S2. Mathematically, this is expressed as:
Cost = ∫(S1 to S2) (M * S^N + C) dS
Solving this integral yields:
Cost = [M / (N + 1) * S^(N + 1) + C * S](S1 to S2)
*/
func (lbc BondingCurve) Cost(x, x1 math.Int) math.Int {
	return lbc.Integral(x1).Sub(lbc.Integral(x))
}

// The Integral of y = M * x^N + C is:
//
//	Cost = (M / (N + 1)) * x^(N + 1) + C * x.
func (lbc BondingCurve) Integral(x math.Int) math.Int {
	// we use osmomath as it support Power function
	xDec := osmomath.BigDecFromSDKDec(ScaleXFromBase(x, lbc.SupplyDecimals()))
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
	integral := xPowNplusOne.Mul(mDivNPlusOne).Add(cx)
	return ScaleDYMToBase(integral.SDKDec())
}

// CalculateM computes the M parameter for a bonding curve
// It's actually not used in the codebase, but it's here for reference and for testing purposes
// val: total value to be raised (in DYM, not adym)
// t: total number of tokens (rollapp's tokens in decimal representation, not base denomination)
// n: curve exponent
// c: constant term
// M = (VAL - C * T) * (N + 1) / T^(N+1)
func CalculateM(val, t, n, c math.LegacyDec) math.LegacyDec {
	valBig := osmomath.BigDecFromSDKDec(val)
	tBig := osmomath.BigDecFromSDKDec(t)
	nBig := osmomath.BigDecFromSDKDec(n)
	cBig := osmomath.BigDecFromSDKDec(c)

	// Calculate N + 1
	nPlusOne := nBig.Add(osmomath.OneDec())

	// Calculate T^(N+1)
	tPowNPlusOne := tBig.Power(nPlusOne)

	// Calculate C * T
	cTimesT := cBig.Mul(tBig)

	// Calculate VAL - C * Z
	numerator := valBig.Sub(cTimesT)

	// Calculate (VAL - C * Z) * (N + 1)
	numerator = numerator.Mul(nPlusOne)

	// Calculate M = numerator / Z^(N+1)
	m := numerator.Quo(tPowNPlusOne)

	// Convert back to math.LegacyDec and return
	return m.SDKDec()
}
