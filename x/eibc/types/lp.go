package types

import (
	"cosmossdk.io/math"
)

func (lp OnDemandLiquidity) MaxSpend() math.Int {
	return math.MinInt(lp.MaxPrice, lp.SpendLimit.Sub(lp.Spent))
}

func (lp OnDemandLiquidity) Accepts(o *DemandOrder) bool {
	priceOK := o.PriceAmount().LTE(lp.MaxSpend())
	feeOK := lp.MinFee.LTE(o.GetFeeAmount())
	return priceOK && feeOK
}
