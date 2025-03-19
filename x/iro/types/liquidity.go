package types

import "cosmossdk.io/math"

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

// find equilibrium amount that will satisfy:
// curve price (x) = balancer pool price (y/x)
// for p(x) = mx^N
// eq = ((N+1) * T) / (R + N + 1)
func FindEquilibrium(curve BondingCurve, totalAllocation math.Int, r math.LegacyDec) math.Int {
	n := curve.N

	// hack for fixed price (as we set N=1 with M=0 instead of N=0)
	if curve.M.IsZero() {
		n = math.LegacyZeroDec()
	}

	n1 := n.Add(math.LegacyOneDec())                         // N + 1
	n2 := n1.Add(r)                                          // N + 1 + R
	eq := (n1.Quo(n2)).MulInt(totalAllocation).TruncateInt() // ((N+1) / (N+1+R)) * T

	return eq
}
