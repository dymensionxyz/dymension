package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (lp OnDemandLiquidity) MaxSpend() math.Int {
	return math.MinInt(lp.MaxPrice, lp.SpendLimit.Sub(lp.Spent))
}

func (lp OnDemandLiquidity) Match(o *eibctypes.DemandOrder) bool {
	priceOK := o.PriceAmount().LTE(lp.MaxSpend())
	feeOK := lp.MinFee.LTE(o.GetFeeAmount())
	return priceOK && feeOK
}
