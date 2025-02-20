package types

import "cosmossdk.io/math"

// CalcLiquidityPoolTokens determines the tokens and DYM to be used for bootstrapping the liquidity pool.
//
// This function calculates the required DYM based on the settled token price and compares it with the raised DYM.
// It returns the amount of RA tokens and DYM to be used for bootstrapping the liquidity pool so it fulfills the last price.
// We expect all the raised dym to be used for liquidity pool, so incentives will consist of the remaining tokens.s
func CalcLiquidityPoolTokens(unsoldRATokens, raisedDYM math.Int, settledTokenPrice math.LegacyDec) (RATokens, dym math.Int) {
	requiredDYM := settledTokenPrice.MulInt(unsoldRATokens).TruncateInt()

	// if raisedDYM is less than requiredDYM, than DYM is the limiting factor
	// we use all the raisedDYM, and the corresponding amount of tokens
	if raisedDYM.LT(requiredDYM) {
		dym = raisedDYM
		RATokens = raisedDYM.ToLegacyDec().Quo(settledTokenPrice).TruncateInt()
	} else {
		// if raisedDYM is more than requiredDYM, than tokens are the limiting factor
		// we use all the unsold tokens, and the corresponding amount of DYM
		RATokens = unsoldRATokens
		dym = requiredDYM
	}

	// for the edge cases where required liquidity truncated to 0
	// we use what we have as it guaranteed to be more than 0
	if dym.IsZero() {
		dym = raisedDYM
	}
	if RATokens.IsZero() {
		RATokens = unsoldRATokens
	}

	return
}

// find equilibrium amount that will satisfy:
// curve price (x) = balancer pool price (y/x)
// for p(x) = mx^N
// eq = ((N+1) / (N+2)) * T
func FindEquilibrium(curve BondingCurve, totalAllocation math.Int) math.Int {
	n := curve.N
	if curve.M.IsZero() {
		n = math.LegacyZeroDec()
	}

	n1 := n.Add(math.LegacyOneDec())                       // N + 1
	n2 := n1.Add(math.LegacyOneDec())                      // N + 2
	eq := n1.Quo(n2).MulInt(totalAllocation).TruncateInt() // ((N+1) / (N+2)) * T

	return eq
}
