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
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid msg")
	}

	if err := k.Slash(ctx, msg.Sequencer, msg.MustRewardee()); err != nil {
		return nil, errorsmod.Wrap(err, "slash sequencer")
	}

	return &types.MsgSlashSequencerResponse{}, nil
}
