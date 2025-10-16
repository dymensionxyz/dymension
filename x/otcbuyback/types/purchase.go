package types

import (
	"time"

	"cosmossdk.io/math"
)

// NewPurchaseEntry creates a new purchase entry
func NewPurchaseEntry(
	amount math.Int,
	vestingStartTime time.Time,
	vestingDuration time.Duration,
) PurchaseEntry {
	return PurchaseEntry{
		Amount:            amount,
		VestingStartTime:  vestingStartTime,
		VestingDuration:   vestingDuration,
	}
}

// NewPurchase creates a new user purchase with empty entries
func NewPurchase() Purchase {
	return Purchase{
		Entries:      []PurchaseEntry{},
		TotalClaimed: math.ZeroInt(),
	}
}

// AddEntry adds a new purchase entry to the user's purchase
func (p *Purchase) AddEntry(entry PurchaseEntry) {
	p.Entries = append(p.Entries, entry)
}

// ClaimTokens updates the purchase data after a successful claim
func (p *Purchase) ClaimTokens(amount math.Int) {
	p.TotalClaimed = p.TotalClaimed.Add(amount)
}

// TotalAmount returns the total amount purchased across all entries
func (p Purchase) TotalAmount() math.Int {
	total := math.ZeroInt()
	for _, entry := range p.Entries {
		total = total.Add(entry.Amount)
	}
	return total
}

// UnclaimedAmount returns the amount still vesting (not yet claimed)
func (p Purchase) UnclaimedAmount() math.Int {
	return p.TotalAmount().Sub(p.TotalClaimed)
}

// CalculateUnlocked calculates the total unlocked amount across all purchase entries
// using the Overlapping approach as described in the ADR
func (p Purchase) CalculateUnlocked(currTime time.Time) math.Int {
	unlocked := math.ZeroInt()

	for _, entry := range p.Entries {
		vestEndTime := entry.VestingStartTime.Add(entry.VestingDuration)

		// If vesting hasn't started yet
		if currTime.Before(entry.VestingStartTime) {
			continue
		}

		// If fully vested
		if currTime.After(vestEndTime) || currTime.Equal(vestEndTime) {
			unlocked = unlocked.Add(entry.Amount)
			continue
		}

		// Partially vested - calculate linear vesting
		elapsed := currTime.Sub(entry.VestingStartTime)
		totalDuration := entry.VestingDuration

		// Calculate vesting progress as a decimal [0 to 1]
		progress := math.LegacyNewDec(elapsed.Nanoseconds()).
			Quo(math.LegacyNewDec(totalDuration.Nanoseconds()))

		// Calculate unlocked amount for this entry
		entryUnlocked := progress.MulInt(entry.Amount).TruncateInt()
		unlocked = unlocked.Add(entryUnlocked)
	}

	return unlocked
}

// ClaimableAmount calculates the amount of tokens that have vested and are claimable
func (p Purchase) ClaimableAmount(currTime time.Time) math.Int {
	unlocked := p.CalculateUnlocked(currTime)
	claimable := unlocked.Sub(p.TotalClaimed)

	// Ensure claimable is not negative
	if claimable.IsNegative() {
		return math.ZeroInt()
	}

	return claimable
}
