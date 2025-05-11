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

// Find the max selling amt such that the price of the liquidity pool is is equal to the last spot price of the bonding curve
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
