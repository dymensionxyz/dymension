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
	AfterUpdateState(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error    // Must be called when a rollapp's state changes
	AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error // Must be called when a rollapp's state changes
	FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error
	RollappCreated(ctx sdk.Context, rollappID, alias string, creator sdk.AccAddress) error
}

var _ RollappHooks = MultiRollappHooks{}

// combine multiple rollapp hooks, all hook functions are run in array sequence

type MultiRollappHooks []RollappHooks

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

func (h MultiRollappHooks) AfterUpdateState(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error {
	for i := range h {
		err := h[i].AfterUpdateState(ctx, rollappID, stateInfo)
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

// RollappCreated implements RollappHooks.
func (h MultiRollappHooks) RollappCreated(ctx sdk.Context, rollappID, alias string, creatorAddr sdk.AccAddress) error {
	for i := range h {
		err := h[i].RollappCreated(ctx, rollappID, alias, creatorAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

type StubRollappCreatedHooks struct{}

func (StubRollappCreatedHooks) RollappCreated(sdk.Context, string, string, sdk.AccAddress) error {
	return nil
}
func (StubRollappCreatedHooks) BeforeUpdateState(sdk.Context, string, string) error      { return nil }
func (StubRollappCreatedHooks) AfterUpdateState(sdk.Context, string, *StateInfo) error   { return nil }
func (StubRollappCreatedHooks) FraudSubmitted(sdk.Context, string, uint64, string) error { return nil }
func (StubRollappCreatedHooks) AfterStateFinalized(sdk.Context, string, *StateInfo) error {
	return nil
}
