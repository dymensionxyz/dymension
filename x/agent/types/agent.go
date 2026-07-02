package types

import "github.com/dymensionxyz/dymension/v3/x/common/tee"

// EffectivePolicy returns the policy in force at the given block height,
// promoting a pending policy once its activation height is reached.
func (a Agent) EffectivePolicy(height int64) tee.Policy {
	if a.PendingPolicy != nil && height >= a.PendingPolicyHeight {
		return *a.PendingPolicy
	}
	return a.Policy
}
