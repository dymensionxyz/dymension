package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// SubmitRollappFraud handles the submission of a fraud proposal
// The fraud proposal can be submitted by the gov module
// We log here, as the error is not bubbled up to the user through the gov proposal
func (k Keeper) SubmitRollappFraud(goCtx context.Context, msg *types.MsgRollappFraudProposal) (*types.MsgRollappFraudProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != k.authority {
		err := errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit fraud proposals")
		ctx.Logger().Error("Fraud proposal.", err)
		return nil, err
	}

	if err := msg.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid msg")
		ctx.Logger().Error("Fraud proposal.", err)
		return nil, err
	}

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		err := errorsmod.Wrap(gerrc.ErrNotFound, "rollapp not found")
		ctx.Logger().Error("Fraud proposal.", err)
		return nil, err
	}

	// check correct revision number (to avoid sending duplicated proposals)
	if rollapp.GetRevisionForHeight(msg.FraudHeight).Number != msg.FraudRevision {
		err := errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "fraud revision number mismatch: %d != %d", rollapp.GetRevisionForHeight(msg.FraudHeight).Number, msg.FraudRevision)
		ctx.Logger().Error("Fraud proposal.", err)
		return nil, err
	}

	// punish the sequencer if needed
	if msg.PunishSequencerAddress != "" {
		err := k.SequencerK.PunishSequencer(ctx, msg.PunishSequencerAddress, msg.MustRewardee())
		if err != nil {
			err = errorsmod.Wrap(err, "jail sequencer")
			ctx.Logger().Error("Fraud proposal.", err)
			return nil, err
		}
	}

	// hard fork the rollapp
	// it will revert the future pending states to the specified height
	// and increment the revision number
	// will fail if state already finalized
	err := k.HardFork(ctx, msg.RollappId, msg.FraudHeight-1)
	if err != nil {
		err = errorsmod.Wrap(err, "hard fork")
		ctx.Logger().Error("Fraud proposal.", err)
		return nil, err
	}

	return &types.MsgRollappFraudProposalResponse{}, nil
}
