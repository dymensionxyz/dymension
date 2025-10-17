package types

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type AuctionStatus string

const (
	AUCTION_STATUS_UPCOMING  AuctionStatus = "upcoming"
	AUCTION_STATUS_ACTIVE    AuctionStatus = "active"
	AUCTION_STATUS_COMPLETED AuctionStatus = "completed"
)

// NewAuction creates a new auction
func NewAuction(
	id uint64,
	allocation math.Int,
	startTime, endTime time.Time,
	discountType DiscountType,
	vestingParams Auction_VestingParams,
	pumpParams Auction_PumpParams,
) Auction {
	return Auction{
		Id:            id,
		Allocation:    allocation,
		StartTime:     startTime,
		EndTime:       endTime,
		DiscountType:  discountType,
		SoldAmount:    math.ZeroInt(),
		VestingParams: vestingParams,
		PumpParams:    pumpParams,
	}
}

// ValidateBasic performs basic validation on the auction
func (a Auction) ValidateBasic() error {
	if a.Id == 0 {
		return ErrInvalidAuctionID
	}

	if !a.Allocation.IsPositive() {
		return ErrInvalidAllocation
	}

	if a.SoldAmount.IsNegative() {
		return ErrInvalidSoldAmount
	}

	if err := a.RaisedAmount.Validate(); err != nil {
		return errors.Join(ErrInvalidRaisedAmount, err)
	}

	// endtime must be greater than starttime
	if a.EndTime.Compare(a.StartTime) <= 0 {
		return ErrInvalidEndTime
	}

	// Validate discount type
	if err := a.DiscountType.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid discount type")
	}

	if a.VestingParams.VestingDelay < 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "vesting delay cannot be negative")
	}

	if a.PumpParams.NumEpochs <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "numEpochsPaidOver must be greater than 0")
	}

	if a.PumpParams.NumOfPumpsPerEpoch <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "numOfPumpsPerEpoch must be greater than 0")
	}
	if a.PumpParams.PumpDelay < 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "pumpDelay cannot be negative")
	}
	if a.PumpParams.PumpInterval <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "pumpInterval must be positive")
	}
	if a.PumpParams.PumpDistr == streamertypes.PumpDistr_PUMP_DISTR_UNSPECIFIED {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "pumpDistr must be specified")
	}

	return nil
}

// GetDiscount returns the discount for a given vesting period
// For LinearDiscount: returns time-based discount (ignores vestingPeriod)
// For FixedDiscount: returns discount for the specified vesting period (ignores currentTime)
func (a Auction) GetDiscount(currentTime time.Time, vestingPeriod time.Duration) (math.LegacyDec, error) {
	switch a.DiscountType.Type.(type) {
	case *DiscountType_Linear:
		return a.DiscountType.GetLinear().GetDiscount(currentTime, a.StartTime, a.EndTime), nil
	case *DiscountType_Fixed:
		return a.DiscountType.GetFixed().GetDiscount(vestingPeriod)
	default:
		return math.LegacyZeroDec(), errors.New("unknown discount type")
	}
}

// GetDiscount calculates discount based on time for LinearDiscount auctions
func (l LinearDiscount) GetDiscount(currentTime, startTime, endTime time.Time) math.LegacyDec {
	// If auction hasn't started, return initial discount
	if currentTime.Before(startTime) {
		return l.InitialDiscount
	}

	// If auction has ended, return max discount
	if currentTime.After(endTime) {
		return l.MaxDiscount
	}

	// Calculate linear progression
	timeElapsed := currentTime.Sub(startTime)
	totalDuration := endTime.Sub(startTime)

	if totalDuration == 0 {
		return l.MaxDiscount
	}

	// Calculate progress as a decimal [0 to 1]
	progress := math.LegacyNewDec(timeElapsed.Nanoseconds()).
		Quo(math.LegacyNewDec(totalDuration.Nanoseconds()))

	// Calculate current discount: initial + (max - initial) * progress
	discountRange := l.MaxDiscount.Sub(l.InitialDiscount)
	return l.InitialDiscount.Add(discountRange.Mul(progress))
}

// GetDiscount returns discount for a specific vesting period in FixedDiscount auctions
func (f FixedDiscount) GetDiscount(vestingPeriod time.Duration) (math.LegacyDec, error) {
	for _, d := range f.Discounts {
		if d.VestingPeriod == vestingPeriod {
			return d.Discount, nil
		}
	}
	return math.LegacyZeroDec(), fmt.Errorf("vesting period not found in auction discount options: %s", vestingPeriod)
}

// GetRemainingAllocation returns the amount of tokens still available for purchase
func (a Auction) GetRemainingAllocation() math.Int {
	return a.Allocation.Sub(a.SoldAmount)
}

// GetVestingStartTime returns the vesting start time for a purchase made at purchaseTime
func (a Auction) GetVestingStartTime(purchaseTime time.Time) time.Time {
	return purchaseTime.Add(a.VestingParams.VestingDelay)
}

/* -------------------------------------------------------------------------- */
/*                              Discount Type                                  */
/* -------------------------------------------------------------------------- */

// Validate performs validation on the DiscountType
func (dt DiscountType) Validate() error {
	switch t := dt.Type.(type) {
	case *DiscountType_Linear:
		return dt.GetLinear().Validate()
	case *DiscountType_Fixed:
		return dt.GetFixed().Validate()
	case nil:
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "discount type must be specified")
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "unknown discount type: %T", t)
	}
}

// Validate performs validation on LinearDiscount
func (ld *LinearDiscount) Validate() error {
	if ld == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "linear discount cannot be nil")
	}

	if ld.InitialDiscount.IsNegative() || ld.InitialDiscount.GTE(math.LegacyOneDec()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "initial discount must be in range [0, 1)")
	}

	if ld.MaxDiscount.IsNegative() || ld.MaxDiscount.GTE(math.LegacyOneDec()) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "max discount must be in range [0, 1)")
	}

	if ld.InitialDiscount.GT(ld.MaxDiscount) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "initial discount must be less than or equal to max discount")
	}

	if ld.VestingPeriod <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "vesting period must be positive")
	}

	return nil
}

// Validate performs validation on FixedDiscount
func (fd *FixedDiscount) Validate() error {
	if fd == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "fixed discount cannot be nil")
	}

	if len(fd.Discounts) <= 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "fixed discount must have at least one discount option")
	}

	seen := make(map[time.Duration]struct{})
	for i, d := range fd.Discounts {
		if d.Discount.IsNegative() || d.Discount.GTE(math.LegacyOneDec()) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "discount %d must be in range [0, 1)", i)
		}

		if d.VestingPeriod <= 0 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "vesting period %d must be positive", i)
		}

		if _, ok := seen[d.VestingPeriod]; ok {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "duplicate vesting period: %s", d.VestingPeriod)
		}
		seen[d.VestingPeriod] = struct{}{}
	}

	return nil
}

/* -------------------------------------------------------------------------- */
/*                                   statuses                                  */
/* -------------------------------------------------------------------------- */

// GetStatus returns the current status of the auction based on time and state
func (a Auction) GetStatus(currentTime time.Time) AuctionStatus {
	if a.Completed {
		return AUCTION_STATUS_COMPLETED
	}

	// Check if auction hasn't started yet
	if currentTime.Before(a.StartTime) {
		return AUCTION_STATUS_UPCOMING
	}

	// Auction is currently active
	return AUCTION_STATUS_ACTIVE
}

// IsActive returns true if the auction is currently active
func (a Auction) IsActive(currentTime time.Time) bool {
	return a.GetStatus(currentTime) == AUCTION_STATUS_ACTIVE
}

// IsCompleted returns true if the auction has completed
func (a Auction) IsCompleted() bool {
	return a.Completed
}

// IsUpcoming returns true if the auction hasn't started yet
func (a Auction) IsUpcoming(currentTime time.Time) bool {
	return a.GetStatus(currentTime) == AUCTION_STATUS_UPCOMING
}
