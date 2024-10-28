package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k msgServer) KickProposer(goCtx context.Context, msg *types.MsgKickProposer) (*types.MsgKickProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.GetRealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	if !seq.Bonded() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "must be bonded to kick")
	}

	proposer := k.GetProposer(ctx, seq.RollappId)
	defer func() {
		k.SetSequencer(ctx, proposer)
	}()

	// TODO: can you ever actually have a situation where the proposer is sentinel and there is a bonded sequencer?
	if !proposer.Sentinel() && k.Kickable(ctx, proposer) {
		if err := k.unbond(ctx, &proposer); err != nil {
			return nil, errorsmod.Wrap(err, "unbond")
		}
		if
		// TODO: also hard fork
	}
	if err := k.ChooseProposer(ctx, seq.RollappId); err != nil {
		return nil, errorsmod.Wrap(err, "choose proposer")
	}
	return &types.MsgKickProposerResponse{}, nil
}
