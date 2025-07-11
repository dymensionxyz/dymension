package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ sequencertypes.Hooks = SequencerHooks{}

type SequencerHooks struct {
	*Keeper
}

// AfterSetRealProposer is called after a new sequencer is set the proposer for a halted rollapp.
// We assume the rollapp had forked once halted
func (h SequencerHooks) AfterSetRealProposer(ctx sdk.Context, rollapp string, newSeq sequencertypes.Sequencer) error {
	// Start the liveness clock from zero
	// NOTE: it could make more sense if liveness was a property of the sequencer rather than the rollapp
	// TODO: tech debt https://github.com/dymensionxyz/dymension/issues/1357

	ra := h.MustGetRollapp(ctx, rollapp)
	h.IndicateLiveness(ctx, &ra)
	h.SetRollapp(ctx, ra)

	// if the rollapp has a state info, set the next proposer to this sequencer
	sInfo, ok := h.GetLatestStateInfo(ctx, rollapp)
	if !ok {
		return nil
	}
	sInfo.NextProposer = newSeq.Address
	h.SetStateInfo(ctx, sInfo)

	return nil
}

// AfterKickProposer is called after a sequencer is kicked from being a proposer.
// We hard fork the rollapp to the latest state so it'll be ready for the next proposer
func (h SequencerHooks) AfterKickProposer(ctx sdk.Context, kicked sequencertypes.Sequencer) error {
	err := h.HardForkToLatest(ctx, kicked.RollappId)
	if err != nil {
		return errorsmod.Wrap(err, "hard fork to latest")
	}
	return nil
}
