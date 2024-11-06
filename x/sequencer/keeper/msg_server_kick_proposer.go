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

	kicker := k.GetSequencer(ctx, msg.GetCreator())
	if !kicker.IsPotentialProposer() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "kicker is not a potential proposer")
	}

	if err := k.Keeper.TryKickProposer(ctx, kicker); err != nil {
		return nil, err
	}

	return &types.MsgKickProposerResponse{}, nil
}
