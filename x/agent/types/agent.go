package types

import (
	"cosmossdk.io/math"

	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// EffectivePolicy returns the policy in force at the given block height,
// promoting a pending policy once its activation height is reached.
func (a Agent) EffectivePolicy(height int64) tee.Policy {
	if a.PendingPolicy != nil && height >= a.PendingPolicyHeight {
		return *a.PendingPolicy
	}
	return a.Policy
}

// SpendEnabled reports whether the agent has a spend policy configured. Empty
// spend_denom means spending is disabled (pure-log agent).
func (a Agent) SpendEnabled() bool {
	return a.SpendDenom != ""
}

// SpendBucket returns the start height of the absolute-aligned tumbling
// window that nowHeight falls into. The window arithmetic mirrors
// x/eibc/types.OnDemandLP (Bucket/RateAllows/RecordSpend).
func (a Agent) SpendBucket(nowHeight uint64) uint64 {
	return nowHeight - (nowHeight % a.SpendWindowBlocks)
}

// spendWindowSpentAmount guards against a nil SpendWindowSpent decoded from
// records written before spending was configured.
func (a Agent) spendWindowSpentAmount() math.Int {
	if a.SpendWindowSpent.IsNil() {
		return math.ZeroInt()
	}
	return a.SpendWindowSpent
}

// RemainingWindowBudget returns the unspent budget of the window containing
// nowHeight. Zero when spending is disabled.
func (a Agent) RemainingWindowBudget(nowHeight uint64) math.Int {
	if !a.SpendEnabled() {
		return math.ZeroInt()
	}
	spent := math.ZeroInt()
	if a.SpendBucket(nowHeight) == a.SpendWindowStartHeight {
		spent = a.spendWindowSpentAmount()
	}
	limit := a.SpendLimitPerWindow
	if limit.IsNil() {
		limit = math.ZeroInt()
	}
	return limit.Sub(spent)
}

// SpendAllows reports whether spending amount at nowHeight stays within the
// per-window cap.
func (a Agent) SpendAllows(nowHeight uint64, amount math.Int) bool {
	return amount.LTE(a.RemainingWindowBudget(nowHeight))
}

// RecordSpend accounts a successful transfer of amount at nowHeight, rolling
// the window over when nowHeight falls into a new bucket.
func (a *Agent) RecordSpend(nowHeight uint64, amount math.Int) {
	if b := a.SpendBucket(nowHeight); b != a.SpendWindowStartHeight {
		a.SpendWindowStartHeight = b
		a.SpendWindowSpent = amount
	} else {
		a.SpendWindowSpent = a.spendWindowSpentAmount().Add(amount)
	}
}
