package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/osmosis-labs/osmosis/osmomath"
)

/*
with the following actions:
  - SpotPrice(x) = M*x^N + C
  - Cost(x, x1) = integral(x1) - integral(x)
    The integral of y = m * x^N + c is (m / (N + 1)) * x^(N + 1) + c * x.
*/

const (
	MaxNValue     = 2
	MaxNPrecision = 3
)

var (
	rollappTokenDefaultDecimals = int64(18) // TODO: allow to be set on creation
	DYMToBaseTokenMultiplier    = math.NewInt(1e18)
)

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

	// Check precision for N
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

// Scales x from it's base denomination to the decimal scale
func scaleXFromBase(x math.Int) math.LegacyDec {
	return math.LegacyNewDecFromIntWithPrec(x, rollappTokenDefaultDecimals)
}

// Scales y from the decimal scale to it's base denomination
func scaleDYMToBase(y math.LegacyDec) math.Int {
	return y.MulInt(DYMToBaseTokenMultiplier).TruncateInt()
}

// SpotPrice returns the spot price at x
func (lbc BondingCurve) SpotPrice(x math.Int) math.Int {
	// we use osmomath as it support Power function
	xDec := osmomath.BigDecFromSDKDec(scaleXFromBase(x))
	nDec := osmomath.BigDecFromSDKDec(lbc.N)
	mDec := osmomath.BigDecFromSDKDec(lbc.M)

	xPowN := xDec.Power(nDec)                    // Calculate x^N
	price := mDec.Mul(xPowN).SDKDec().Add(lbc.C) // M * x^N + C

	return scaleDYMToBase(price)
}

// Cost returns the cost of buying x1 - x tokens
func (lbc BondingCurve) Cost(x, x1 math.Int) math.Int {
	return lbc.Integral(x1).Sub(lbc.Integral(x))
}

// The Integral of y = M * x^N + C is:
//
//	(M / (N + 1)) * x^(N + 1) + C * x.
func (lbc BondingCurve) Integral(x math.Int) math.Int {
	// we use osmomath as it support Power function
	xDec := osmomath.BigDecFromSDKDec(scaleXFromBase(x))
	mDec := osmomath.BigDecFromSDKDec(lbc.M)
	cDec := osmomath.BigDecFromSDKDec(lbc.C)
	nPlusOne := osmomath.BigDecFromSDKDec(lbc.N.Add(math.LegacyNewDec(1)))

	xPowNplusOne := xDec.Power(nPlusOne)  // Calculate x^(N + 1)
	mDivNPlusOne := mDec.QuoMut(nPlusOne) // Calculate m / (N + 1)
	cx := cDec.Mul(xDec)                  // Calculate C * x

	// Calculate the integral
	integral := xPowNplusOne.Mul(mDivNPlusOne).Add(cx)
	return scaleDYMToBase(integral.SDKDec())
}

// CalculateM computes the M parameter for a bonding curve
// It's actually not used in the codebase, but it's here for reference and for testing purposes
// val: total value to be raised (in DYM, not adym)
// t: total number of tokens (rollapp's tokens in decimal scale, not base denomination)
// n: curve exponent
// c: constant term
// M = (VAL - C * T) * (N + 1) / T^(N+1)
func CalculateM(val, t, n, c math.LegacyDec) math.LegacyDec {
	// Convert to osmomath.BigDec for more precise calculations
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
