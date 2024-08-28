package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
)

/*
with the following actions:
  - SpotPrice(x) = M*x^N + C
  - Cost(x, x1) = integral(x1) - integral(x)
    The integral of y = m * x^N + c is (m / (N + 1)) * x^(N + 1) + c * x.
*/
type BondingCurve struct {
	//FIXME: change to Dec
	M math.Int
	N uint64
	C math.Int
}

// NewBondingCurve creates a new  bonding curve
func NewBondingCurve(m, c, n math.Int) BondingCurve {
	return BondingCurve{
		M: m,
		C: c,
		N: n.Uint64(),
	}
}

// validateBasic checks if the bonding curve is valid
func (lbc BondingCurve) ValidateBasic() error {
	if lbc.M.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "m: %s", lbc.M.String())
	}
	if !lbc.C.IsPositive() {
		return errorsmod.Wrapf(ErrInvalidBondingCurve, "c: %s", lbc.C.String())
	}
	return nil
}

// SpotPrice returns the spot price at x
func (lbc BondingCurve) SpotPrice(x math.Int) math.Int {
	xN := x.ToLegacyDec().Power(lbc.N).TruncateInt() // Calculate x^N
	return xN.Mul(lbc.M).Add(lbc.C)
}

// Cost returns the cost of buying x1 - x tokens
func (lbc BondingCurve) Cost(x, x1 math.Int) math.Int {
	return lbc.Integral(x1).Sub(lbc.Integral(x))
}

// The Integral of y = M * x^N + C is:
//
//	(M / (N + 1)) * x^(N + 1) + C * x.
func (lbc BondingCurve) Integral(x math.Int) math.Int {
	nPlusOne := int64(lbc.N + 1)

	xNPlusOne := x.ToLegacyDec().Power(uint64(nPlusOne))              // Calculate x^(N + 1)
	mDivNPlusOne := lbc.M.ToLegacyDec().QuoInt(math.NewInt(nPlusOne)) // Calculate m / (N + 1)
	cx := lbc.C.Mul(x)                                                // Calculate C * x

	// Calculate the integral
	integral := xNPlusOne.Mul(mDivNPlusOne).TruncateInt().Add(cx)
	return integral
}
