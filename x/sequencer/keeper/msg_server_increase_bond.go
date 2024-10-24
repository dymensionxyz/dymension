package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.validateBondDenom(ctx, msg.AddAmount); err != nil {
		return nil, err
	}

	seq, err := k.tryGetSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if err := k.sendToModule(ctx, &seq, msg.AddAmount); err != nil {
		return nil, err
	}

	// TODO: write seq

	// emit a typed event which includes the added amount and the active bond amount
	return &types.MsgIncreaseBondResponse{}, uevent.EmitTypedEvent(ctx,
		&types.EventIncreasedBond{
			Sequencer:   msg.Creator,
			Bond:        seq.Tokens,
			AddedAmount: msg.AddAmount,
		},
	)
}
