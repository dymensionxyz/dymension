package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// These can be utilized to communicate between a rollapp keeper and another
// keeper which must take particular actions when rollapp change state.
// The second keeper must implement this interface, which then the
// rollapp keeper can call.

// RollappHooks event hooks for rollapp object (noalias)
type RollappHooks interface {
	BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) // Must be called when a rollapp's state changes
}

var _ RollappHooks = MultiRollappHooks{}

// combine multiple rollapp hooks, all hook functions are run in array sequence
type MultiRollappHooks []RollappHooks

// Creates hooks for the Rollapp Module.
func NewMultiRollappHooks(hooks ...RollappHooks) MultiRollappHooks {
	return hooks
}

func (h MultiRollappHooks) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) {
	for i := range h {
		h[i].BeforeUpdateState(ctx, seqAddr, rollappId)
	}
}
