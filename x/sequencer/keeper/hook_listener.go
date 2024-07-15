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

func (hook rollappHook) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
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

	if !sequencer.Proposer {
		return types.ErrNotActiveSequencer
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
		err := hook.k.forceUnbondSequencer(ctx, sequencer.SequencerAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

// RollappCreated implements types.RollappHooks.
func (hook rollapphook) RollappCreated(ctx sdk.Context, rollappID string) error {
	return nil
}
