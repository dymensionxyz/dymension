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

	if !seq.IsPotentialProposer() {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not ready to propose")
	}

	proposer := k.GetProposer(ctx, seq.RollappId)

	if !proposer.Sentinel() && k.Kickable(ctx, proposer) {
		if err := k.unbond(ctx, &proposer); err != nil {
			return nil, errorsmod.Wrap(err, "unbond")
		}
		k.SetSequencer(ctx, proposer)
		k.optOutAllSequencers(ctx, seq.RollappId, seq.Address)
		// TODO: also hard fork
	}
	if err := k.ChooseProposer(ctx, seq.RollappId); err != nil {
		return nil, errorsmod.Wrap(err, "choose proposer")
	}
	return &types.MsgKickProposerResponse{}, nil
}
