package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.TryGetSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "try get seq")
	}
	err = k.tryUnbond(ctx, seq, nil)
	if errorsmod.IsOf(types.ErrUnbondProposerOrSuccessor) {
		completionTime := k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &completionTime,
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &types.MsgUnbondResponse{}, nil
}
