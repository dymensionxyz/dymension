package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// TryKickProposer tries to remove the incumbent proposer. It requires the incumbent
// proposer to be below a threshold of bond. The caller must also be bonded and opted in.
func (k msgServer) KickProposer(goCtx context.Context, msg *types.MsgKickProposer) (*types.MsgKickProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	kicker := k.GetSequencer(ctx, msg.GetCreator())
	if !kicker.IsPotentialProposer() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "kicker is not a potential proposer")
	}

	ra := kicker.RollappId

	proposer := k.GetProposer(ctx, ra)

	if !k.Kickable(ctx, proposer) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not kickable")
	}

	if err := k.abruptRemoveSequencer(ctx, proposer); err != nil {
		return nil, errorsmod.Wrap(err, "force remove sequencer")
	}

	if err := kicker.SetOptedIn(ctx, true); err != nil {
		return nil, errorsmod.Wrap(err, "set opted in")
	}
	k.SetSequencer(ctx, kicker)

	if err := k.RecoverFromSentinel(ctx, ra); err != nil {
		return nil, errorsmod.Wrap(err, "recover from sentinel")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventKickedProposer{
		Rollapp:  ra,
		Kicker:   kicker.Address,
		Proposer: proposer.Address,
	}); err != nil {
		return nil, err
	}

	return &types.MsgKickProposerResponse{}, nil
}
