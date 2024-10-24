package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.tryGetSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	err = k.tryUnbond(ctx, &seq, seq.TokensCoin())
	if errorsmod.IsOf(types.ErrUnbondProposerOrSuccessor) {
		completionTime := k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &completionTime,
			},
		}, nil
	}
	if err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	// TODO: write seq
	return &types.MsgUnbondResponse{}, nil
}
