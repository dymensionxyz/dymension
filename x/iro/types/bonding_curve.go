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
	MaxNValue    = 10
	MaxPrecision = 2
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

	// Check precision for M, N, and C
	if !checkPrecision(lbc.M) || !checkPrecision(lbc.N) || !checkPrecision(lbc.C) {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "m, n, and c must have at most %d decimal places", MaxPrecision)
	}

	return nil
}

// checkPrecision checks if a math.LegacyDec has at most MaxPrecision decimal places
func checkPrecision(d math.LegacyDec) bool {
	// Multiply by 10^MaxPrecision and check if it's an integer
	multiplied := d.Mul(math.LegacyNewDec(10).Power(uint64(MaxPrecision)))
	return multiplied.IsInteger()
}

// SpotPrice returns the spot price at x
func (lbc BondingCurve) SpotPrice(x math.Int) math.LegacyDec {
	// we use osmomath as it support Power function
	xDec := osmomath.BigDecFromSDKDec(x.ToLegacyDec())
	nDec := osmomath.BigDecFromSDKDec(lbc.N)
	mDec := osmomath.BigDecFromSDKDec(lbc.M)

	xPowN := xDec.Power(nDec)                  // Calculate x^N
	return mDec.Mul(xPowN).SDKDec().Add(lbc.C) // M * x^N + C
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
	xDec := osmomath.BigDecFromSDKDec(x.ToLegacyDec())
	mDec := osmomath.BigDecFromSDKDec(lbc.M)
	cDec := osmomath.BigDecFromSDKDec(lbc.C)
	nPlusOne := osmomath.BigDecFromSDKDec(lbc.N.Add(math.LegacyNewDec(1)))

	xPowNplusOne := xDec.Power(nPlusOne)  // Calculate x^(N + 1)
	mDivNPlusOne := mDec.QuoMut(nPlusOne) // Calculate m / (N + 1)
	cx := cDec.Mul(xDec)                  // Calculate C * x

	// Calculate the integral
	integral := xPowNplusOne.Mul(mDivNPlusOne).Add(cx)
	return integral.SDKDec().TruncateInt()
}
