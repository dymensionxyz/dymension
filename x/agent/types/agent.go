package types

import (
	"bytes"
	"math/big"

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
// still inside the trailing window: prev * (W - elapsed) / W. The product is
// formed in an unbounded big.Int because prev may be as large as the limit,
// and math.Int panics on intermediates beyond 256 bits; the quotient is at most
// prev, so it always fits. Floor division under-counts the weighted value by
// less than one base unit.
func PrevWindowWeighted(prev math.Int, elapsed, windowBlocks uint64) math.Int {
	n := new(big.Int).Mul(prev.BigInt(), new(big.Int).SetUint64(windowBlocks-elapsed))
	return math.NewIntFromBigInt(n.Quo(n, new(big.Int).SetUint64(windowBlocks)))
}

// RemainingWindowBudget returns the unspent budget at nowHeight under a
// weighted two-bucket sliding window: limit minus current bucket spend minus
// the prorated previous bucket spend. Zero when spending is disabled.
//
// This closes the tumbling-window burst at the bucket edge (a full previous
// bucket leaves zero budget at the next bucket start), but it is an
// approximation that assumes previous-bucket spend was spread evenly. Spend
// concentrated late in the previous bucket is under-weighted, so actual spend
// over an arbitrary W-block interval can approach 2x the limit: with W=100 and
// limit=1000, 1000 at height 199 and 980 at height 298 are both allowed. That
// approximation error is separate from, and far larger than, the sub-unit
// floor rounding in PrevWindowWeighted.
func (a Agent) RemainingWindowBudget(nowHeight uint64) math.Int {
	if !a.SpendEnabled() {
		return math.ZeroInt()
	}
	bucket, cur, prev := a.spendWindow(nowHeight)
	weighted := PrevWindowWeighted(prev, nowHeight-bucket, a.SpendWindowBlocks)
	limit := a.SpendLimitPerWindow
	if limit.IsNil() {
		limit = math.ZeroInt()
	}
	// Subtract one term at a time and saturate: cur + weighted may each be up
	// to the limit, so their sum can exceed math.Int's 256-bit range.
	remaining := limit.Sub(cur)
	if remaining.LTE(weighted) {
		return math.ZeroInt()
	}
	return remaining.Sub(weighted)
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
