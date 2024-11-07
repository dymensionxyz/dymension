package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) KickProposer(goCtx context.Context, msg *types.MsgKickProposer) (*types.MsgKickProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.RealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.TryKickProposer(ctx, seq); err != nil {
		return nil, err
	}

	return &types.MsgKickProposerResponse{}, nil
}
