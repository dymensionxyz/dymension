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

	seq, ok := k.GetSequencer(ctx, msg.Creator)
	if !ok {
		return nil, types.ErrSequencerNotFound
	}

	if !seq.Bonded() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "must be bonded to kick")
	}

	proposer, err := k.GetProposer(ctx, seq.RollappId)
	if err != nil {
		return nil, err
	}
	kickThreshold := k.GetParams(ctx).KickThreshold
	if proposer.TokensCoin().IsLT(kickThreshold) {
		k.unbond(ctx, proposer)
		k.chooseProposer(ctx)
		// TODO: also hard fork
	}
	return &types.MsgKickProposerResponse{}, nil
}
