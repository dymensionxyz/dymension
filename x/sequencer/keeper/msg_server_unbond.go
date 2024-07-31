package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	if !seq.IsBonded() {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not bonded: got %s",
			seq.Status.String(),
		)
	}

	if seq.UnbondRequestHeight != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer has already requested to unbond",
		)
	}
	seq.UnbondRequestHeight = ctx.BlockHeight()

	// check if the sequencer is required for a notice period before unbonding
	if k.isNoticePeriodRequired(ctx, seq) {
		completionTime := k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &completionTime,
			},
		}, nil
	}

	// otherwise, start unbonding
	completionTime := k.startUnbondingPeriodForSequencer(ctx, &seq)
	return &types.MsgUnbondResponse{
		CompletionTime: &types.MsgUnbondResponse_UnbondingCompletionTime{
			UnbondingCompletionTime: &completionTime,
		},
	}, nil
}
