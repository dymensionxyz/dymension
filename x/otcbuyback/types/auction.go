package types

import (
	"time"

	"cosmossdk.io/math"
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
	initialDiscount, maxDiscount math.LegacyDec,
	vestingPlan Auction_VestingParams,
) Auction {

	return Auction{
		Id:              id,
		Allocation:      allocation,
		StartTime:       startTime,
		EndTime:         endTime,
		InitialDiscount: initialDiscount,
		MaxDiscount:     maxDiscount,
		SoldAmount:      math.ZeroInt(),
		VestingParams: Auction_VestingParams{
			VestingPeriod:               vestingPlan.VestingPeriod,
			VestingStartAfterAuctionEnd: vestingPlan.VestingStartAfterAuctionEnd,
		},
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

	if a.EndTime.Before(a.StartTime) {
		return ErrInvalidEndTime
	}

	if a.InitialDiscount.IsNegative() || a.InitialDiscount.GTE(math.LegacyOneDec()) {
		return ErrInvalidDiscount
	}

	if a.MaxDiscount.IsNegative() || a.MaxDiscount.GTE(math.LegacyOneDec()) {
		return ErrInvalidDiscount
	}

	if a.InitialDiscount.GT(a.MaxDiscount) {
		return ErrInvalidDiscount
	}

	if a.VestingParams.VestingPeriod <= 0 {
		return ErrVestingParam
	}

	if a.VestingParams.VestingStartAfterAuctionEnd <= 0 {
		return ErrVestingParam
	}

	return nil
}

// GetCurrentDiscount calculates the current discount percentage based on time elapsed
func (a Auction) GetCurrentDiscount(currentTime time.Time) math.LegacyDec {
	// If auction hasn't started, return initial discount
	if currentTime.Before(a.StartTime) {
		return a.InitialDiscount
	}

	// If auction has ended, return max discount
	if currentTime.After(a.EndTime) {
		return a.MaxDiscount
	}

	// Calculate linear progression
	timeElapsed := currentTime.Sub(a.StartTime)
	totalDuration := a.EndTime.Sub(a.StartTime)

	// Calculate progress as a decimal (0 to 1)
	progress := math.LegacyNewDec(timeElapsed.Nanoseconds()).
		Quo(math.LegacyNewDec(totalDuration.Nanoseconds()))

	// Ensure progress is between 0 and 1
	if progress.GT(math.LegacyOneDec()) {
		progress = math.LegacyOneDec()
	}
	if progress.LT(math.LegacyZeroDec()) {
		progress = math.LegacyZeroDec()
	}

	// Calculate current discount: initial + (max - initial) * progress
	discountRange := a.MaxDiscount.Sub(a.InitialDiscount)
	currentDiscount := a.InitialDiscount.Add(discountRange.Mul(progress))

	return currentDiscount
}

// GetRemainingAllocation returns the amount of tokens still available for purchase
func (a Auction) GetRemainingAllocation() math.Int {
	return a.Allocation.Sub(a.SoldAmount)
}

// GetVestingStartTime returns the start time of the vesting period
func (a Auction) GetVestingStartTime() time.Time {
	return a.EndTime.Add(a.VestingParams.VestingStartAfterAuctionEnd)
}

// GetVestingEndTime returns the end time of the vesting period
func (a Auction) GetVestingEndTime() time.Time {
	return a.GetVestingStartTime().Add(a.VestingParams.VestingPeriod)
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
