package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ sequencertypes.Hooks = SequencerHooks{}

type SequencerHooks struct {
	*Keeper
}

func (k SequencerHooks) AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after sequencertypes.Sequencer) {
	// Start the liveness clock from zero
	// NOTE: it could make more sense if liveness was a property of the sequencer rather than the rollapp
	// TODO: tech debt https://github.com/dymensionxyz/dymension/issues/1357

	ra := k.MustGetRollapp(ctx, rollapp)
	k.ResetLivenessClock(ctx, &ra)
	if !after.Sentinel() {
		k.ScheduleLivenessEvent(ctx, &ra)
	}
	k.SetRollapp(ctx, ra)
}

func (s SequencerHooks) AfterKickProposer(ctx sdk.Context, kicked sequencertypes.Sequencer) {
	// TODO: trigger a hard fork
}
