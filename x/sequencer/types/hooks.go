package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Hooks interface {
	AfterRecoveryFromHalt(ctx sdk.Context, rollapp string, newProposer Sequencer) error
	AfterKickProposer(ctx sdk.Context, rollapp string) error
}

var _ Hooks = NoOpHooks{}

type NoOpHooks struct{}

func (n NoOpHooks) AfterRecoveryFromHalt(ctx sdk.Context, rollapp string, newProposer Sequencer) error {
	return nil
}

func (n NoOpHooks) AfterKickProposer(ctx sdk.Context, rollapp string) error {
	return nil
}

var _ Hooks = MultiHooks{}

type MultiHooks []Hooks

func NewMultiHooks(hooks ...Hooks) MultiHooks {
	return MultiHooks(hooks)
}

func (m MultiHooks) AfterRecoveryFromHalt(ctx sdk.Context, rollapp string, newProposer Sequencer) error {
	for _, h := range m {
		err := h.AfterRecoveryFromHalt(ctx, rollapp, newProposer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m MultiHooks) AfterKickProposer(ctx sdk.Context, rollapp string) error {
	for _, h := range m {
		err := h.AfterKickProposer(ctx, rollapp)
		if err != nil {
			return err
		}
	}
	return nil
}
