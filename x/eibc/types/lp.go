package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (r OnDemandLPRecord) Validate() error {
	if err := r.Lp.Validate(); err != nil {
		return errorsmod.Wrap(err, "base")
	}
	if r.Spent.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "negative spent")
	}
	if r.Spent.GT(r.Lp.SpendLimit) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spent greater than spend limit")
	}
	return nil
}

func (r OnDemandLP) Validate() error {
	// TODO:
	return nil
}

func (r OnDemandLPRecord) MaxSpend() math.Int {
	return math.MinInt(r.Lp.MaxPrice, r.Lp.SpendLimit.Sub(r.Spent))
}

func (r OnDemandLPRecord) Accepts(o *DemandOrder) bool {
	priceOK := o.PriceAmount().LTE(r.MaxSpend())
	feeOK := r.Lp.MinFee.LTE(o.GetFeeAmount())
	return priceOK && feeOK
}
