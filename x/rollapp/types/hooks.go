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
	BeforeUpdateState(ctx sdk.Context, seqAddr, rollappId string, lastStateUpdateBySequencer bool) error // Must be called when a rollapp's state changes
	AfterUpdateState(ctx sdk.Context, stateInfo *StateInfoMeta) error                                    // Must be called when a rollapp's state changes
	AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *StateInfo) error                   // Must be called when a rollapp's state changes
	RollappCreated(ctx sdk.Context, rollappID, alias string, creator sdk.AccAddress) error
	AfterTransfersEnabled(ctx sdk.Context, rollappID, rollappIBCDenom string) error

	OnHardFork(ctx sdk.Context, rollappID string, height uint64) error
}

var _ RollappHooks = MultiRollappHooks{}

// combine multiple rollapp hooks, all hook functions are run in array sequence

type MultiRollappHooks []RollappHooks

func NewMultiRollappHooks(hooks ...RollappHooks) MultiRollappHooks {
	return hooks
}

func (h MultiRollappHooks) BeforeUpdateState(ctx sdk.Context, seqAddr, rollappId string, lastStateUpdateBySequencer bool) error {
	for i := range h {
		err := h[i].BeforeUpdateState(ctx, seqAddr, rollappId, lastStateUpdateBySequencer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h MultiRollappHooks) AfterUpdateState(ctx sdk.Context, stateInfo *StateInfoMeta) error {
	for i := range h {
		err := h[i].AfterUpdateState(ctx, stateInfo)
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

func (h MultiRollappHooks) OnHardFork(ctx sdk.Context, rollappID string, lastValidHeight uint64) error {
	for i := range h {
		err := h[i].OnHardFork(ctx, rollappID, lastValidHeight)
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

func (h MultiRollappHooks) AfterTransfersEnabled(ctx sdk.Context, rollappID, rollappIBCDenom string) error {
	for i := range h {
		err := h[i].AfterTransfersEnabled(ctx, rollappID, rollappIBCDenom)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ RollappHooks = &StubRollappCreatedHooks{}

type StubRollappCreatedHooks struct{}

func (StubRollappCreatedHooks) RollappCreated(sdk.Context, string, string, sdk.AccAddress) error {
	return nil
}

func (StubRollappCreatedHooks) BeforeUpdateState(sdk.Context, string, string, bool) error { return nil }
func (StubRollappCreatedHooks) AfterUpdateState(sdk.Context, *StateInfoMeta) error {
	return nil
}
func (StubRollappCreatedHooks) OnHardFork(sdk.Context, string, uint64) error { return nil }
func (StubRollappCreatedHooks) AfterStateFinalized(sdk.Context, string, *StateInfo) error {
	return nil
}

func (StubRollappCreatedHooks) AfterTransfersEnabled(sdk.Context, string, string) error { return nil }
