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
// We log here, as the error is not bubbled up to the user through the gov proposal
func (k Keeper) SubmitRollappFraud(goCtx context.Context, msg *types.MsgRollappFraudProposal) (*types.MsgRollappFraudProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != k.authority {
		err := errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit fraud proposals")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	if err := msg.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid msg")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		err := errorsmod.Wrap(gerrc.ErrNotFound, "rollapp not found")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	// check revision number
	if rollapp.RevisionNumber != msg.RollappRevision {
		err := errorsmod.Wrap(gerrc.ErrFailedPrecondition, "revision number mismatch")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	// validate the rollapp is past its genesis bridge phase
	if !rollapp.IsTransferEnabled() {
		err := errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp is not past genesis bridge phase")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	// punish the sequencer if needed
	if msg.PunishSequencerAddress != "" {
		err := k.sequencerKeeper.PunishSequencer(ctx, msg.PunishSequencerAddress, msg.MustRewardee())
		if err != nil {
			err = errorsmod.Wrap(err, "jail sequencer")
			ctx.Logger().Error(err.Error())
			return nil, err
		}
	}

	// hard fork the rollapp
	// it will revert the future pending states to the specified height
	// and increment the revision number
	// will fail if state already finalized
	err := k.HardFork(ctx, msg.RollappId, msg.FraudHeight)
	if err != nil {
		err = errorsmod.Wrap(err, "hard fork")
		ctx.Logger().Error(err.Error())
		return nil, err
	}

	return &types.MsgRollappFraudProposalResponse{}, nil
}
