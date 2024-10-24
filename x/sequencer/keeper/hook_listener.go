package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

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
		return types.ErrNotProposer
	}

	if lastStateUpdateBySequencer {
		return errorsmod.Wrap(hook.k.onProposerLastBlock(ctx, proposer), "on proposer last block")
	}

	return nil
}
