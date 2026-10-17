package types

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// PromotePendingPolicy applies a matured pending policy in place, so
// verification at or after the activation height uses the rotated policy.
// Callers persist the agent to make the promotion durable.
func (a *Agent) PromotePendingPolicy(height int64) {
	if a.PendingPolicy != nil && height >= a.PendingPolicyHeight {
		a.Policy = *a.PendingPolicy
		a.PendingPolicy = nil
		a.PendingPolicyHeight = 0
	}
}

// EffectivePolicy returns the policy in force at height without mutating the
// stored agent.
func (a Agent) EffectivePolicy(height int64) tee.Policy {
	a.PromotePendingPolicy(height)
	return a.Policy
}

// SpendEnabled reports whether the agent has a spend policy configured. Empty
// spend_denom means spending is disabled (pure-log agent).
func (a Agent) SpendEnabled() bool {
	return a.SpendDenom != ""
}

// SpendRecipientAllowed reports whether recipient may receive an attested
// transfer. An empty allowlist means unrestricted (spending is still bounded
// by denom + per-window limit).
func (a Agent) SpendRecipientAllowed(recipient string) bool {
	if len(a.SpendRecipientAllowlist) == 0 {
		return true
	}
	address, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return false
	}
	for _, r := range a.SpendRecipientAllowlist {
		allowedAddress, err := sdk.AccAddressFromBech32(r)
		if err == nil && bytes.Equal(allowedAddress, address) {
			return true
		}
	}
	return false
}

// ValidateSpendState checks that the agent's spend policy and window
// bookkeeping are a state the runtime could have produced, guarding genesis
// imports against e.g. an enabled agent with a zero window length (which would
// make SpendBucket divide by zero).
func (a Agent) ValidateSpendState(spendRecipientAllowlistMax uint64) error {
	if err := ValidateSpendPolicy(a.SpendDenom, a.SpendLimitPerWindow, a.SpendWindowBlocks); err != nil {
		return err
	}
	if uint64(len(a.SpendRecipientAllowlist)) > spendRecipientAllowlistMax {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend recipient allowlist exceeds max")
	}
	recipients := make(map[string]struct{}, len(a.SpendRecipientAllowlist))
	for _, recipient := range a.SpendRecipientAllowlist {
		address, err := sdk.AccAddressFromBech32(recipient)
		if err != nil {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend recipient allowlist address")
		}
		key := string(address)
		if _, duplicate := recipients[key]; duplicate {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "duplicate spend recipient allowlist address")
		}
		recipients[key] = struct{}{}
	}
	spent := a.spendWindowSpentAmount()
	if spent.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "negative spend window spent")
	}
	prevSpent := a.spendPrevWindowSpentAmount()
	if prevSpent.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "negative spend previous window spent")
	}
	if !a.SpendEnabled() {
		if !spent.IsZero() || !prevSpent.IsZero() || a.SpendWindowStartHeight != 0 {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend window state set without spend denom")
		}
		return nil
	}
	if spent.GT(a.SpendLimitPerWindow) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend window spent greater than spend limit")
	}
	if prevSpent.GT(a.SpendLimitPerWindow) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend previous window spent greater than spend limit")
	}
	// RecordSpend only ever writes bucket starts (multiples of the window
	// length), so anything else is not a state the runtime can produce.
	if a.SpendWindowStartHeight%a.SpendWindowBlocks != 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend window start height not bucket-aligned")
	}
	return nil
}

// SpendBucket returns the start height of the absolute-aligned bucket that
// nowHeight falls into. Buckets feed the weighted sliding-window estimate in
// spendWindow.
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

// spendPrevWindowSpentAmount guards against a nil SpendPrevWindowSpent decoded
// from records written before the field existed. Zero means "no previous-window
// spend", which never grants extra budget, so no state migration is needed.
func (a Agent) spendPrevWindowSpentAmount() math.Int {
	if a.SpendPrevWindowSpent.IsNil() {
		return math.ZeroInt()
	}
	return a.SpendPrevWindowSpent
}

// spendWindow migrates the stored window bookkeeping to the bucket containing
// nowHeight, returning that bucket and its current and previous bucket spend.
func (a Agent) spendWindow(nowHeight uint64) (bucket uint64, cur, prev math.Int) {
	bucket = a.SpendBucket(nowHeight)
	switch a.SpendWindowStartHeight {
	case bucket:
		return bucket, a.spendWindowSpentAmount(), a.spendPrevWindowSpentAmount()
	case bucket - a.SpendWindowBlocks:
		return bucket, math.ZeroInt(), a.spendWindowSpentAmount()
	default:
		return bucket, math.ZeroInt(), math.ZeroInt()
	}
}

// PrevWindowWeighted prorates the previous bucket's spend by the fraction of it
// still inside the trailing window: prev * (W - elapsed) / W. Integer floor
// division keeps it deterministic; it under-counts by less than one base unit,
// negligible next to the 2x boundary burst a tumbling window allows.
func PrevWindowWeighted(prev math.Int, elapsed, windowBlocks uint64) math.Int {
	return prev.Mul(math.NewIntFromUint64(windowBlocks - elapsed)).Quo(math.NewIntFromUint64(windowBlocks))
}

// RemainingWindowBudget returns the unspent budget at nowHeight under a
// weighted sliding window: current bucket spend plus the prorated previous
// bucket spend, so spend over any trailing window length stays at roughly the
// limit instead of doubling across a bucket edge. Zero when spending is
// disabled.
func (a Agent) RemainingWindowBudget(nowHeight uint64) math.Int {
	if !a.SpendEnabled() {
		return math.ZeroInt()
	}
	bucket, cur, prev := a.spendWindow(nowHeight)
	spent := cur.Add(PrevWindowWeighted(prev, nowHeight-bucket, a.SpendWindowBlocks))
	limit := a.SpendLimitPerWindow
	if limit.IsNil() {
		limit = math.ZeroInt()
	}
	return math.MaxInt(math.ZeroInt(), limit.Sub(spent))
}

// SpendAllows reports whether spending amount at nowHeight stays within the
// per-window cap.
func (a Agent) SpendAllows(nowHeight uint64, amount math.Int) bool {
	return amount.LTE(a.RemainingWindowBudget(nowHeight))
}

// RecordSpend accounts a successful transfer of amount at nowHeight, first
// migrating the window bookkeeping to nowHeight's bucket.
func (a *Agent) RecordSpend(nowHeight uint64, amount math.Int) {
	bucket, cur, prev := a.spendWindow(nowHeight)
	a.SpendWindowStartHeight = bucket
	a.SpendPrevWindowSpent = prev
	a.SpendWindowSpent = cur.Add(amount)
}
