package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) SubmitSlashProposal(goCtx context.Context, msg *types.MsgSlashSequencerProposal) (_ *types.MsgSlashSequencerResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	defer func() {
		if err != nil {
			ctx.Logger().Error("Submit slash proposal", err)
		}
	}()

	// check if the proposer is the authority
	if msg.Authority != k.authority {
		err = errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the authority can submit slash proposals")
		return
	}

	// validate the message
	if err = msg.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "msg")
		return
	}

	var seq types.Sequencer
	seq, err = k.RealSequencer(ctx, msg.Sequencer)
	if err != nil {
		return
	}

	if err = k.SlashAllTokens(ctx, msg.Sequencer, msg.MustRewardee()); err != nil {
		return
	}

	err = k.abruptRemoveSequencer(ctx, seq)
	if err != nil {
		return nil, errorsmod.Wrap(err, "abrupt remove sequencer")
	}

	return &types.MsgSlashSequencerResponse{}, nil
}
