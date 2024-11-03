package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// FraudProposalHandler handles the submission of a fraud proposal
// The fraud proposal can be submitted by the gov module
func (k Keeper) FraudProposalHandler(ctx sdk.Context, msg types.MsgRollappFraudProposal) error {
	if msg.Authority != k.authority {
		return errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit fraud proposals")
	}

	if err := msg.ValidateBasic(); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid msg")
	}

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return errorsmod.Wrap(gerrc.ErrNotFound, "rollapp not found")
	}
	// check revision number
	if rollapp.RevisionNumber != msg.RollappRevision {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "revision number mismatch")
	}

	// validate the fraud height is already committed
	sinfo, found := k.GetLatestStateInfo(ctx, msg.RollappId)
	if !found || sinfo.GetLatestHeight() < msg.FraudHeight {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "fraud height not committed")
	}

	// check wether the fraud height is already finalized
	sinfo, found = k.GetLatestFinalizedStateInfo(ctx, msg.RollappId)
	if found && sinfo.GetLatestHeight() >= msg.FraudHeight {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "fraud height already finalized")
	}

	// jail the sequencer if needed
	if msg.SlashSequencerAddress != "" {
		err := k.sequencerKeeper.JailByAddr(ctx, msg.SlashSequencerAddress)
		if err != nil {
			return errorsmod.Wrap(err, "jail sequencer")
		}
	}
	return k.HardFork(ctx, msg.RollappId, msg.FraudHeight)
}
