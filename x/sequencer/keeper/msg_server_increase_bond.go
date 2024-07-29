package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// IncreaseBond implements types.MsgServer.
func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the sequencer already exists
	sequencer, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	// check if the sequencer is bonded
	if !sequencer.IsBonded() {
		return nil, types.ErrInvalidSequencerStatus
	}

	// check if sequencer is currently jailed
	if sequencer.Jailed {
		return nil, types.ErrSequencerJailed
	}

	// transfer the bond from the sequencer to the module account
	seqAcc, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.Amount))
	if err != nil {
		return nil, err
	}

	// update the sequencers bond amount
	sequencer.Tokens = sequencer.Tokens.Add(msg.Amount)
	k.UpdateSequencer(ctx, sequencer, sequencer.Status)

	// emit an event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBondIncreased,
			sdk.NewAttribute(types.AttributeKeySequencer, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBond, sequencer.Tokens.String()),
		),
	)

	return &types.MsgIncreaseBondResponse{}, nil
}
