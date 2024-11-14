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
func (k Keeper) SubmitRollappFraud(goCtx context.Context, msg *types.MsgRollappFraudProposal) (_ *types.MsgRollappFraudProposalResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	defer func() {
		if err != nil {
			ctx.Logger().Error("Submit rollapp fraud.", err)
		}
	}()

	if msg.Authority != k.authority {
		err = errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit fraud proposals")
		return
	}

	if err = msg.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "msg")
		return
	}

	// punish the sequencer if needed
	if msg.PunishSequencerAddress != "" {
		err = k.sequencerKeeper.SlashAllTokens(ctx, msg.PunishSequencerAddress, msg.MustRewardee())
		if err != nil {
			err = errorsmod.Wrap(err, "jail sequencer")
			return
		}
	}

	// hard fork the rollapp
	// it will revert the future pending states to the specified height
	// and increment the revision number
	// will fail if state already finalized
	err = k.HardFork(ctx, msg.RollappId, msg.FraudHeight)
	if err != nil {
		err = errorsmod.Wrap(err, "hard fork")
		return
	}

	return &types.MsgRollappFraudProposalResponse{}, nil
}
