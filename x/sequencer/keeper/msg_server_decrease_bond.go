package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// DecreaseBond implements types.MsgServer.
func (k msgServer) DecreaseBond(goCtx context.Context, msg *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sequencer, err := k.bondUpdateAllowed(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	// Check if the sequencer has enough bond to decrease
	if sequencer.Tokens.IsAllLTE(sdk.NewCoins(msg.DecreaseAmount)) {
		return nil, types.ErrInsufficientBond
	}

	if _, found := k.getSequencerDecreasingBond(ctx, msg.Creator); found {
		return nil, types.ErrBondDecreaseInProgress
	}

	// Check if the bond reduction will make the sequencer's bond less than the minimum bond value
	minBondValue := k.GetParams(ctx).MinBond
	if !minBondValue.IsNil() && !minBondValue.IsZero() {
		decreasedBondValue := sequencer.Tokens.Sub(msg.DecreaseAmount)
		if decreasedBondValue.IsAllLT(sdk.NewCoins(minBondValue)) {
			return nil, types.ErrInsufficientBond
		}
	}
	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	k.setDecreasingBondQueue(ctx, types.BondReduction{
		SequencerAddress: msg.Creator,
		UnbondAmount:     msg.DecreaseAmount,
		UnbondTime:       completionTime,
	})

	return &types.MsgDecreaseBondResponse{
		CompletionTime: completionTime,
	}, nil
}
