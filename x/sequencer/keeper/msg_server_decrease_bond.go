package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// DecreaseBond implements types.MsgServer.
func (k msgServer) DecreaseBond(goCtx context.Context, msg *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the sequencer from the store
	sequencer, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	// Check if the sequencer is currently bonded
	if !sequencer.IsBonded() {
		return nil, types.ErrInvalidSequencerStatus
	}

	// // check if sequencer is currently jailed
	// if sequencer.Jailed {
	// 	return nil, types.ErrSequencerJailed
	// }

	// Check if the bond reduction will make the sequencer's bond less than the minimum bond value
	minBondValue := k.GetParams(ctx).MinBond
	if !minBondValue.IsNil() && !minBondValue.IsZero() {
		decreasedBondValue := sequencer.Tokens.Sub(msg.Amount)
		if decreasedBondValue.IsAllLT(sdk.NewCoins(minBondValue)) {
			return nil, types.ErrInsufficientBond
		}
	}

	//k.setdecreasingbondsequencerqueue(ctx, sequencer, msg.Amount)
	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))

	return &types.MsgDecreaseBondResponse{
		CompletionTime: completionTime,
	}, nil
}
