package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
)

var _ rollapptypes.RollappHooks = rollapphook{}

// Hooks wrapper struct for rollapp keeper.
type rollapphook struct {
	k Keeper
}

// Return the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollapphook{k}
}

func (hook rollapphook) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) {
	// fmt.Printf("BeforeUpdateState seqAddr(%s), rollappId(%s)\n", seqAddr, rollappId)
	// hook.k.Logger(ctx).Error(fmt.Sprintf("not implemented: BeforeUpdateState seqAddr(%s), rollappId(%s)\n", seqAddr, rollappId))

	// check to see if the sequencer has been registered before
	sequencer, found := hook.k.GetSequencer(ctx, seqAddr)
	if !found {
		panic(types.ErrUnknownSequencer)
	}

	// check to see if the rollappId matches the one of the sequencer
	if sequencer.RollappId != rollappId {
		panic(types.ErrSequencerRollappMismatch)
	}

	// check to see if the sequencer is active and can make the update
	scheduler, found := hook.k.GetScheduler(ctx, seqAddr)
	if !found {
		panic(sdkerrors.Wrapf(sdkerrors.ErrLogic, "sequencer address: %s not registered in scheduler", seqAddr))
	}
	if scheduler.Status != types.Proposer {
		panic(types.ErrNotActiveSequencer)
	}

}
