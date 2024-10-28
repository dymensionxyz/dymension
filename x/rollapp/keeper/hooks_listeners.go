package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ sequencertypes.Hooks = SequencerHooks{}

type SequencerHooks struct {
	*Keeper
}

func (s SequencerHooks) AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after sequencertypes.Sequencer) {
	// if after is not sentinel, reschedule
	s.DelLivenessEvents(ctx, rollapp)
	if !after.Sentinel() {
		s.RestartLivenessClock(ctx)
	}
	// TODO implement me
	panic("implement me")
}

func (s SequencerHooks) AfterKickProposer(ctx sdk.Context, kicked sequencertypes.Sequencer) {
	// TODO implement me
	panic("implement me")
}
