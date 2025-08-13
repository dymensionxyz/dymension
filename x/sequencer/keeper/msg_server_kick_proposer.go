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

	kicker, err := k.RealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	// Prevent self-kick: kicker cannot be the current proposer
	proposer := k.GetProposer(ctx, kicker.RollappId)
	if kicker.Address == proposer.Address {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "sequencer cannot kick itself")
	}

	if err := k.TryKickProposer(ctx, kicker); err != nil {
		return nil, err
	}

	return &types.MsgKickProposerResponse{}, nil
}
