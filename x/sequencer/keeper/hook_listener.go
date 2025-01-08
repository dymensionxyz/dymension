package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ rollapptypes.RollappHooks = rollappHook{}

type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{k: k}
}

// BeforeUpdateState will reject if the caller is not proposer
// if lastStateUpdateBySequencer is true, validate that the sequencer is in the middle of a rotation
func (hook rollappHook) BeforeUpdateState(ctx sdk.Context, seqAddr, rollappId string, lastStateUpdateBySequencer bool) error {
	proposer := hook.k.GetProposer(ctx, rollappId)
	if seqAddr != proposer.Address {
		return types.ErrNotProposer
	}

	// if lastStateUpdateBySequencer is true, validate that the sequencer is in the middle of a rotation
	if lastStateUpdateBySequencer && !hook.k.AwaitingLastProposerBlock(ctx, rollappId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "sequencer is not in the middle of a rotation")
	}

	return nil
}

// AfterUpdateState checks if rotation is completed and the nextProposer is changed
func (hook rollappHook) AfterUpdateState(ctx sdk.Context, stateInfo *rollapptypes.StateInfoMeta) error {
	proposer := hook.k.GetProposer(ctx, stateInfo.Rollapp)
	return hook.k.afterStateUpdate(ctx, proposer, stateInfo.Sequencer != stateInfo.NextProposer)
}

// OnHardFork implements the RollappHooks interface
// unbonds all rollapp sequencers
// slashing / jailing is handled by the caller, outside of this function
func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappID string, _ uint64) error {
	err := hook.k.optOutAllSequencers(ctx, rollappID)
	if err != nil {
		return errorsmod.Wrap(err, "opt out all sequencers")
	}

	// clear current proposer and successor
	hook.k.abruptRemoveProposer(ctx, rollappID)
	hook.k.SetSuccessor(ctx, rollappID, types.SentinelSeqAddr)

	return nil
}
