package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (d OnDemandLP) Addr() (sdk.AccAddress, error) {
	return sdk.AccAddressFromBech32(d.FundsAddr)
}

func (d OnDemandLP) MustAddr() sdk.AccAddress {
	a, err := d.Addr()
	if err != nil {
		panic(err)
	}
	return a
}

func (d OnDemandLP) Validate() error {
	if _, err := d.Addr(); err != nil {
		return errorsmod.Wrap(err, "addr")
	}
	if err := validateRollappID(d.Rollapp); err != nil {
		return errorsmod.Wrap(err, "rollapp id")
	}
	if sdk.ValidateDenom(d.Denom) != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "denom")
	}
	if d.MaxPrice.IsNil() || !d.MaxPrice.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "max price")
	}
	if d.MinFee.IsNil() || d.MinFee.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min fee")
	}
	if d.SpendLimit.IsNil() || !d.SpendLimit.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend limit")
	}
	return nil
}

func (r OnDemandLPRecord) Validate() error {
	if r.Lp == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty lp")
	}
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

func (r OnDemandLPRecord) MaxSpend() math.Int {
	return math.MinInt(r.Lp.MaxPrice, r.Lp.SpendLimit.Sub(r.Spent))
}

func (r OnDemandLPRecord) Accepts(nowHeight uint64, o commontypes.DemandOrder) bool {
	priceOK := o.PriceAmount().LTE(r.MaxSpend())
	feeOK := r.Lp.MinFee.LTE(o.GetFeeAmount())
	ageOK := r.Lp.OrderMinAgeBlocks <= nowHeight-o.CreationHeight
	return priceOK && feeOK && ageOK
}
