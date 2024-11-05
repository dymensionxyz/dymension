package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ sequencertypes.Hooks = SequencerHooks{}

type SequencerHooks struct {
	*Keeper
}

func (h SequencerHooks) AfterChooseNewProposer(ctx sdk.Context, rollapp string, before, after sequencertypes.Sequencer) {
	// Start the liveness clock from zero
	// NOTE: it could make more sense if liveness was a property of the sequencer rather than the rollapp
	// TODO: tech debt https://github.com/dymensionxyz/dymension/issues/1357

	ra := h.Keeper.MustGetRollapp(ctx, rollapp)
	h.Keeper.ResetLivenessClock(ctx, &ra)
	if !after.Sentinel() {
		h.Keeper.ScheduleLivenessEvent(ctx, &ra)
	}
	h.Keeper.SetRollapp(ctx, ra)

	// recover from halt
	// if the rollapp has a state info, set the next proposer to this sequencer
	if before.Sentinel() && !after.Sentinel() {
		sInfo, _ := h.Keeper.GetLatestStateInfo(ctx, rollapp)
		sInfo.NextProposer = after.Address
		h.Keeper.SetStateInfo(ctx, sInfo)
	}

}

func (h SequencerHooks) AfterKickProposer(ctx sdk.Context, kicked sequencertypes.Sequencer) {
	h.Keeper.HardForkToLatest(ctx, kicked.RollappId)
}
