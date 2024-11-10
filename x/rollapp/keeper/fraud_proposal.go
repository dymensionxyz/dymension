package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// SubmitRollappFraud handles the submission of a fraud proposal
// The fraud proposal can be submitted by the gov module
func (k Keeper) SubmitRollappFraud(goCtx context.Context, msg *types.MsgRollappFraudProposal) (*types.MsgRollappFraudProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit fraud proposals")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid msg")
	}

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, errorsmod.Wrap(gerrc.ErrNotFound, "rollapp not found")
	}
	// check revision number
	if rollapp.RevisionNumber != msg.RollappRevision {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "revision number mismatch")
	}

	// validate the rollapp is past it's genesis bridge phase
	if !rollapp.IsTransferEnabled() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp is not past genesis bridge phase")
	}

	// validate we have state infos committed
	sinfo, found := k.GetLatestStateInfo(ctx, msg.RollappId)
	if !found {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no state info found")
	}

	// check whether the fraud height is already finalized
	sinfo, found = k.GetLatestFinalizedStateInfo(ctx, msg.RollappId)
	if found && sinfo.GetLatestHeight() >= msg.FraudHeight {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "fraud height already finalized")
	}

	// punish the sequencer if needed
	if msg.PunishSequencerAddress != "" {
		err := k.sequencerKeeper.PunishSequencer(ctx, msg.PunishSequencerAddress, msg.MustRewardee())
		if err != nil {
			return nil, errorsmod.Wrap(err, "jail sequencer")
		}
	}

	err := k.HardFork(ctx, msg.RollappId, msg.FraudHeight)
	if err != nil {
		return nil, errorsmod.Wrap(err, "hard fork")
	}

	return &types.MsgRollappFraudProposalResponse{}, nil
}
