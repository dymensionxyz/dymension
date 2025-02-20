package types

import (
	"cosmossdk.io/math"
)

func (r OnDemandLPRecord) MaxSpend() math.Int {
	return math.MinInt(r.Lp.MaxPrice, r.Lp.SpendLimit.Sub(r.Spent))
}

func (r OnDemandLPRecord) Accepts(o *DemandOrder) bool {
	priceOK := o.PriceAmount().LTE(r.MaxSpend())
	feeOK := r.Lp.MinFee.LTE(o.GetFeeAmount())
	return priceOK && feeOK
}
