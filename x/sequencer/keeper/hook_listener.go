package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ rollapptypes.RollappHooks = rollapphook{}

// Hooks wrapper struct for rollapp keeper.
type rollapphook struct {
	rollapptypes.BaseRollappHook
	k Keeper
}

// Return the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollapphook{
		rollapptypes.BaseRollappHook{},
		k,
	}
}

func (hook rollapphook) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
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
	if sequencer.Status != types.Proposer {
		return types.ErrNotActiveSequencer
	}
	return nil
}
