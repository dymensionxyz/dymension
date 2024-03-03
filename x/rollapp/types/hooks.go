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
	BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error         // Must be called when a rollapp's state changes
	AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error // Must be called when a rollapp's state changes
	FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error
}

var _ RollappHooks = MultiRollappHooks{}

// combine multiple rollapp hooks, all hook functions are run in array sequence
type MultiRollappHooks []RollappHooks

// Creates hooks for the Rollapp Module.
func NewMultiRollappHooks(hooks ...RollappHooks) MultiRollappHooks {
	return hooks
}

func (h MultiRollappHooks) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	for i := range h {
		err := h[i].BeforeUpdateState(ctx, seqAddr, rollappId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h MultiRollappHooks) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error {
	for i := range h {
		err := h[i].AfterStateFinalized(ctx, rollappID, stateInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h MultiRollappHooks) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	for i := range h {
		err := h[i].FraudSubmitted(ctx, rollappID, height, seqAddr)
		if err != nil {
			return err
		}
	}
	return nil
}
