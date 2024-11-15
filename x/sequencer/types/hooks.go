package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Hooks interface {
	AfterSetRealProposer(ctx sdk.Context, rollapp string, newProposer Sequencer) error
	AfterKickProposer(ctx sdk.Context, kicked Sequencer) error
}

var _ Hooks = NoOpHooks{}

type NoOpHooks struct{}

func (n NoOpHooks) AfterSetRealProposer(ctx sdk.Context, rollapp string, newProposer Sequencer) error {
	return nil
}

func (n NoOpHooks) AfterKickProposer(ctx sdk.Context, kicked Sequencer) error {
	return nil
}

var _ Hooks = MultiHooks{}

type MultiHooks []Hooks

func NewMultiHooks(hooks ...Hooks) MultiHooks {
	return MultiHooks(hooks)
}

func (m MultiHooks) AfterSetRealProposer(ctx sdk.Context, rollapp string, newProposer Sequencer) error {
	for _, h := range m {
		err := h.AfterSetRealProposer(ctx, rollapp, newProposer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m MultiHooks) AfterKickProposer(ctx sdk.Context, kicked Sequencer) error {
	for _, h := range m {
		err := h.AfterKickProposer(ctx, kicked)
		if err != nil {
			return err
		}
	}
	return nil
}
