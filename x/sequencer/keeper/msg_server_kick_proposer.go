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

	seq, err := k.tryGetSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if !seq.Bonded() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "must be bonded to kick")
	}

	proposer := k.GetProposer(ctx, seq.RollappId)
	kickThreshold := k.GetParams(ctx).KickThreshold
	// TODO: can you ever actually have a situation where the proposer is sentinel and there is a bonded sequencer?
	if !proposer.Sentinel() && proposer.TokensCoin().IsLT(kickThreshold) {
		if err := k.unbond(ctx, &proposer); err != nil {
			return nil, errorsmod.Wrap(err, "unbond")
		}
		// TODO: also hard fork
	}
	if err := k.chooseProposer(ctx, seq.RollappId); err != nil {
		return nil, errorsmod.Wrap(err, "choose proposer")
	}
	return &types.MsgKickProposerResponse{}, nil
}
