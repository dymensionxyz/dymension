package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Hooks interface {
	AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after Sequencer)
	AfterKickProposer(ctx sdk.Context, kicked Sequencer) error
}

var _ Hooks = NoOpHooks{}

type NoOpHooks struct{}

func (n NoOpHooks) AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after Sequencer) {
}

func (n NoOpHooks) AfterKickProposer(ctx sdk.Context, kicked Sequencer) error {
	return nil
}

var _ Hooks = MultiHooks{}

type MultiHooks []Hooks

func NewMultiHooks(hooks ...Hooks) MultiHooks {
	return MultiHooks(hooks)
}

func (m MultiHooks) AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after Sequencer) {
	for _, h := range m {
		h.AfterChooseNewProposer(ctx, rollapp, before, after)
	}
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
