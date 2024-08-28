package keeper

import (
	"cosmossdk.io/math"
)

/*
linear bonding curve is of the form y = m * x + c

with the following actions:
- SpotPrice(x) = m * x + c
- Cost(x, x1) = integral(x1) - integral(x)
- integral(x) = 0.5 * m * x^2 + c * x
*/
type LinearBondingCurve struct {
	M math.Int
	C math.Int
}

// NewLinearBondingCurve creates a new linear bonding curve
func NewLinearBondingCurve(m, c math.Int) LinearBondingCurve {
	return LinearBondingCurve{
		M: m,
		C: c,
	}
}

// SpotPrice returns the spot price at x
func (lbc LinearBondingCurve) SpotPrice(x math.Int) math.Int {
	return x.Mul(lbc.M).Add(lbc.C)
}

// Cost returns the cost of buying x1 - x tokens
func (lbc LinearBondingCurve) Cost(x, x1 math.Int) math.Int {
	return lbc.integral(x1).Sub(lbc.integral(x))
}

func (lbc LinearBondingCurve) integral(x math.Int) math.Int {
	// Calculate mx^2/2
	mx := x.Mul(x).Mul(lbc.M).QuoRaw(2)
	// Calculate cx
	cx := lbc.C.Mul(x)

	// Sum the parts
	return mx.Add(cx)
}
