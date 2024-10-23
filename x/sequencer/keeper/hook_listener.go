package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

// Hooks wrapper struct for rollapp keeper.
type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

// RollappHooks returns the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{k: k}
}

// BeforeUpdateState checks various conditions before updating the state.
// It verifies if the sequencer has been registered, if the rollappId matches the one of the sequencer,
// if there is a proposer for the given rollappId, and if the sequencer is the active one.
// If the lastStateUpdateBySequencer flag is true, it also checks if the rollappId is rotating and
// performs a rotation of the proposer.
// Returns an error if any of the checks fail, otherwise returns nil.
func (hook rollappHook) BeforeUpdateState(ctx sdk.Context, seqAddr, rollappId string, lastStateUpdateBySequencer bool) error {
	proposer := hook.k.GetProposer(ctx, rollappId)
	if seqAddr != proposer.Address {
		return gerrc.ErrInternal.Wrap("got state update from non-proposer")
	}

	if lastStateUpdateBySequencer {
		return errorsmod.Wrap(hook.k.onProposerLastBlock(ctx, proposer), "on proposer last block")
	}

	return nil
}
