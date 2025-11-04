package types

import (
	"errors"

	"cosmossdk.io/math"
	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	MaxFindGraduationIterations = 100
	RollappDenomDecimals        = 18
)

// CalcLiquidityPoolTokens determines the tokens and liquidity to be used for bootstrapping the liquidity pool.
//
// This function calculates the required liquidity based on the settled token price and compares it with the raised liquidity.
// It returns the amount of RA tokens and liquidity to be used for bootstrapping the liquidity pool so it fulfills the last price.
// We expect all the raised liquidity to be used for liquidity pool, so incentives will consist of the remaining tokens.s
func CalcLiquidityPoolTokens(unsoldRATokens, raisedLiquidity math.Int, settledTokenPrice math.LegacyDec) (RATokens, liquidity math.Int) {
	requiredLiquidity := settledTokenPrice.MulInt(unsoldRATokens).TruncateInt()

	// if raisedLiquidity is less than requiredLiquidity, than liquidity is the limiting factor
	// we use all the raisedLiquidity, and the corresponding amount of tokens
	if raisedLiquidity.LT(requiredLiquidity) {
		liquidity = raisedLiquidity
		RATokens = raisedLiquidity.ToLegacyDec().Quo(settledTokenPrice).TruncateInt()
	} else {
		// if raisedLiquidity is more than requiredLiquidity, than tokens are the limiting factor
		// we use all the unsold tokens, and the corresponding amount of liquidity
		RATokens = unsoldRATokens
		liquidity = requiredLiquidity
	}

	// for the edge cases where required liquidity truncated to 0
	// we use what we have as it guaranteed to be more than 0
	if liquidity.IsZero() {
		liquidity = raisedLiquidity
	}
	if RATokens.IsZero() {
		RATokens = unsoldRATokens
	}

	return
}

// Find the max selling amt such that the price of the liquidity pool is equal to the last spot price of the bonding curve
//
// Assuming c=0 (enforced in curve validation):
//
//	    Define SpotIRO(x)=mx^n
//	    Define RaisedLiquidity(x)=(mx^(n+1))/(n+1) [by integral]
//	    Define SpotPool(x)=(r*RaisedLiquidity(x))/(totalAllocation-x)  [x is sold amt]
//	    Solve SpotIRO=SpotPool [cancel x^n terms and rearrange linear eq]
//	=> x=((n+1)*totalAllocation)/(r+n+1)
//
// If c!=0 and m=0:
//
//		SpotIRO(x)=c
//		RaisedLiquidity(x)=cx
//		SpotPool(x)=(r*cx)/(totalAllocation-x)
//		Solve SpotIRO=SpotPool [cancel c terms and rearrange linear eq]
//	 => x=totalAllocation/(r+1) [same as above calculation but n=0]
func FindEquilibrium(curve BondingCurve, totalAllocation math.Int, r math.LegacyDec) math.Int {
	n := curve.N

	if curve.M.IsZero() { // c is allowed to be non-zero
		n = math.LegacyZeroDec()
	}

	n1 := n.Add(math.LegacyOneDec())                         // N + 1
	n2 := n1.Add(r)                                          // N + 1 + R
	eq := (n1.Quo(n2)).MulInt(totalAllocation).TruncateInt() // ((N+1) / (N+1+R)) * T

	return eq
}

// graduationTargetG returns G(x) as LegacyDec, where
//
//	G(x) = C*x * [ ((T-2x) / ((N+1)*( x*(N+2)/(N+1) - T ))) + 1 ]
//
// This is algebraically equal to L(x) under the fair-seeding constraint.
// We want G(x) == VAL/2.
func graduationTargetG(
	x math.LegacyDec,
	T math.LegacyDec,
	N math.LegacyDec,
	C math.LegacyDec,
) math.LegacyDec {
	one := math.LegacyOneDec()

	nPlusOne := N.Add(one)        // N+1
	nPlusTwo := nPlusOne.Add(one) // N+2

	// (T - 2x)
	tMinus2x := T.Sub(x.MulInt64(2))

	// x * (N+2)/(N+1) - T
	frac := nPlusTwo.Quo(nPlusOne) // (N+2)/(N+1)
	xFrac := x.Mul(frac)
	denomInner := xFrac.Sub(T)

	// denomFull = (N+1) * ( x*(N+2)/(N+1) - T )
	denomFull := nPlusOne.Mul(denomInner)

	// termA = (T-2x)/denomFull
	termA := tMinus2x.Quo(denomFull)

	// bracket = termA + 1
	bracket := termA.Add(one)

	// G(x) = C * x * bracket
	return C.Mul(x).Mul(bracket)
}

// MOfX returns M(x) for debugging / downstream use:
//
//	M(x) = C*(T-2x) / [ x^N * ( x*(N+2)/(N+1) - T ) ]
//
// Assumes x is in valid band so denominator != 0.
func MOfX(
	x math.LegacyDec,
	T math.LegacyDec,
	N math.LegacyDec,
	C math.LegacyDec,
) math.LegacyDec {
	one := math.LegacyOneDec()
	nPlusOne := N.Add(one)
	nPlusTwo := nPlusOne.Add(one)

	// numerator = C * (T - 2x)
	num := C.Mul(T.Sub(x.MulInt64(2)))

	// x^N using BigDec for proper handling of decimal exponents
	xBig := osmomath.BigDecFromSDKDec(x)
	nBig := osmomath.BigDecFromSDKDec(N)

	xPowN := xBig.Power(nBig) // Calculate x^N

	// inner = x*(N+2)/(N+1) - T
	frac := nPlusTwo.Quo(nPlusOne) // (N+2)/(N+1)
	inner := x.Mul(frac).Sub(T)

	denom := xPowN.SDKDec().Mul(inner)
	return num.Quo(denom)
}

// FindGraduation finds x_g (graduation point) given:
// - totalAllocation T (in scaled/decimal form, e.g. 1e9 for 1 billion tokens)
// - curve exponent N
// - floor price C (price per token in scaled form)
// - target valuation VAL (in scaled/decimal form, e.g. 40000 for 40K)
//
// Returns the graduation point in BASE DENOMINATION.
//
// It uses the condition:
//
//	L(x_g) = VAL/2
//
// where
//
//	L(x) = (M/(N+1))*x^(N+1) + C*x
//
// and M is eliminated using the fair-seeding constraint
//
//	L(x_g) = (T - x_g) * P(x_g), P(x)=M*x^N + C
//
// We binary search x in (T/2 , T*(N+1)/(N+2)) and stop
// when the interval is <=1% of T. We then return the best x in Int.

// The limits are reached by:
// - lowLimit (T/2) - for fixed price curve (N=0)
// - highLimit (T*(N+1)/(N+2)) - for C=0 curves
func FindGraduation(
	T math.LegacyDec,
	N math.LegacyDec,
	C math.LegacyDec,
	VAL math.LegacyDec,
) (math.LegacyDec, error) {
	if T.IsZero() {
		return math.LegacyZeroDec(), errors.New("allocation is zero")
	}
	if C.IsZero() {
		return math.LegacyZeroDec(), errors.New("curve C is zero")
	}
	if N.IsNegative() {
		return math.LegacyZeroDec(), errors.New("curve N is negative")
	}
	// Preconditions we assume are already validated externally:
	// - T > 0
	// - C > 0
	// - N >= 1 and integral (N == floor(N))
	// - VAL > 0

	// Compute search interval:
	//
	// valid band for positive M is:
	//   x_low  > T/2
	//   x_high < T * (N+1)/(N+2)
	//
	// We'll clamp to ints and keep it closed.

	// low = floor(T/2)
	lowLimit := T.QuoInt64(2)

	// high = floor( T * (N+1)/(N+2) ) - 1
	nPlusOne := N.Add(math.LegacyOneDec())        // N+1
	nPlusTwo := nPlusOne.Add(math.LegacyOneDec()) // N+2
	ratio := nPlusOne.Quo(nPlusTwo)               // (N+1)/(N+2)
	highLimit := T.Mul(ratio)

	// target = VAL/2
	target := VAL.QuoInt64(2)

	// 0.1% precision threshold = ceil( VAL / 1000 )
	// we'll compute once
	threshold := VAL.QuoInt64(1000)

	// helper: compare G(x) vs target.
	cmp := func(x math.LegacyDec) int {
		gx := graduationTargetG(x, T, N, C) // LegacyDec

		// Check if |gx - target| <= threshold
		diff := gx.Sub(target).Abs()
		if diff.LTE(threshold) {
			return 0 // within threshold
		}

		// gx ? target
		if gx.GT(target) {
			return 1
		}
		return -1
	}

	lowX := lowLimit
	highX := highLimit

	count := 0

	for {
		mid := lowX.Add(highX).QuoInt64(2)
		count++

		switch cmp(mid) {
		case -1:
			// G(mid) < target → need to sell more → move low up
			lowX = mid
		case 1:
			// G(mid) > target → too high valuation → move high down
			highX = mid
		default:
			// perfect hit
			return mid, nil
		}

		if count > MaxFindGraduationIterations {
			return math.LegacyZeroDec(), errors.New("max iterations reached")
		}
	}
}
