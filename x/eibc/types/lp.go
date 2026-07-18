package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	if d.MinFee.IsNil() || d.MinFee.IsNegative() || d.MinFee.GT(math.LegacyNewDec(1)) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min fee")
	}
	if d.SpendLimit.IsNil() || !d.SpendLimit.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend limit")
	}
	if !d.MinFeeAbsolute.IsNil() && d.MinFeeAbsolute.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min fee absolute must be non-negative")
	}
	if d.rateLimitEnabled() {
		if !d.RateLimitAmount.IsPositive() {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "rate limit amount must be positive")
		}
		if d.RateLimitBlocks == 0 {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "rate limit amount set without rate limit blocks")
		}
	} else if d.RateLimitBlocks != 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "rate limit blocks set without rate limit amount")
	}
	return nil
}

// rateLimitEnabled reports whether the velocity cap is configured. A nil or
// non-positive amount means disabled, matching the opt-in semantics.
func (d OnDemandLP) rateLimitEnabled() bool {
	return !d.RateLimitAmount.IsNil() && d.RateLimitAmount.IsPositive()
}

// minFeeAbsolute guards against a nil read from pre-upgrade LP records; nil or
// zero means the absolute fee floor is disabled.
func (d OnDemandLP) minFeeAbsolute() math.Int {
	if d.MinFeeAbsolute.IsNil() {
		return math.ZeroInt()
	}
	return d.MinFeeAbsolute
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

// Bucket returns the start height of the absolute-aligned tumbling window that
// nowHeight falls into.
func (d OnDemandLP) Bucket(nowHeight uint64) uint64 {
	return nowHeight - (nowHeight % d.RateLimitBlocks)
}

// windowSpentAmount guards against a nil WindowSpent read from pre-upgrade
// records, mirroring how Spent is treated.
func (r OnDemandLPRecord) windowSpentAmount() math.Int {
	if r.WindowSpent.IsNil() {
		return math.ZeroInt()
	}
	return r.WindowSpent
}

// RateAllows reports whether spending price at nowHeight stays within the
// velocity cap. Disabled rate limiting always allows.
func (r OnDemandLPRecord) RateAllows(nowHeight uint64, price math.Int) bool {
	if !r.Lp.rateLimitEnabled() {
		return true
	}
	spent := math.ZeroInt()
	if r.Lp.Bucket(nowHeight) == r.WindowStartHeight {
		spent = r.windowSpentAmount()
	}
	return price.LTE(r.Lp.RateLimitAmount.Sub(spent))
}

// RecordSpend accounts a successful fill of price at nowHeight, updating both
// the cumulative Spent and the rate window. Window bookkeeping is skipped when
// rate limiting is disabled (RateLimitBlocks == 0, which would divide by zero).
func (r *OnDemandLPRecord) RecordSpend(nowHeight uint64, price math.Int) {
	r.Spent = r.Spent.Add(price)
	if !r.Lp.rateLimitEnabled() {
		return
	}
	if b := r.Lp.Bucket(nowHeight); b != r.WindowStartHeight {
		r.WindowStartHeight = b
		r.WindowSpent = price
	} else {
		r.WindowSpent = r.windowSpentAmount().Add(price)
	}
}

func (r OnDemandLPRecord) Accepts(nowHeight uint64, o *DemandOrder) bool {
	priceOK := o.EffectivePriceAmount(nowHeight).LTE(r.MaxSpend())
	feeOK := r.Lp.MinFee.LTE(o.EffectiveFeePercent(nowHeight))
	minFeeAbsOK := o.EffectiveFeeAmount(nowHeight).GTE(r.Lp.minFeeAbsolute())
	ageOK := r.Lp.OrderMinAgeBlocks <= nowHeight-o.CreationHeight
	validOK := r.Lp.ValidUntilHeight == 0 || nowHeight < r.Lp.ValidUntilHeight
	rateOK := r.RateAllows(nowHeight, o.PriceAmount())
	return priceOK && feeOK && minFeeAbsOK && ageOK && validOK && rateOK
}
