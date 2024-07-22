package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

// Hooks wrapper struct for rollapp keeper.
type rollappHook struct {
	k Keeper
}

// RollappHooks returns the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{
		k,
	}
}

func (hook rollappHook) BeforeUpdateState(ctx sdk.Context, seqAddr, rollappId string, lastStateOfSequencer bool) error {
	// check to see if the sequencer has been registered before
	sequencer, found := hook.k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	// check to see if the rollappId matches the one of the sequencer
	if sequencer.RollappId != rollappId {
		return types.ErrSequencerRollappMismatch
	}

	// check to see if the sequencer is active and can make the update
	if sequencer.Status != types.Bonded {
		return types.ErrInvalidSequencerStatus
	}

	seq, ok := hook.k.GetProposer(ctx, rollappId)
	if !ok {
		return types.ErrNoProposer
	}
	if sequencer.SequencerAddress != seq.SequencerAddress {
		return types.ErrNotActiveSequencer
	}

	if lastStateOfSequencer {
		if !hook.k.IsRotating(ctx, rollappId) {
			return types.ErrInvalidRequest
		}
		// TODO: the hub should probably validate the lastBlock in the lastBatch,
		// to make sure the sequencer is passing the correct nextSequencer on the L2

		hook.k.RotateProposer(ctx, rollappId)
	}

	return nil
}

func (hook rollappHook) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	return nil
}

// FraudSubmitted implements the RollappHooks interface
// It slashes the sequencer and unbonds all other bonded sequencers
func (hook rollappHook) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	err := hook.k.Slashing(ctx, seqAddr)
	if err != nil {
		return err
	}

	// unbond all other bonded sequencers
	sequencers := hook.k.GetSequencersByRollappByStatus(ctx, rollappID, types.Bonded)
	for _, sequencer := range sequencers {
		hook.k.startUnbondingPeriodForSequencer(ctx, &sequencer)
	}

	// clear the proposer and next proposer
	hook.k.RemoveActiveSequencer(ctx, rollappID)
	hook.k.RemoveNextProposer(ctx, rollappID)

	return nil
}
