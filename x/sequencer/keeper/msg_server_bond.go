package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.validBondDenom(ctx, msg.AddAmount); err != nil {
		return nil, err
	}

	seq, err := k.GetRealSequencer(ctx, msg.GetCreator())
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

func (k msgServer) DecreaseBond(goCtx context.Context, msg *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.GetRealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}
	if err := k.tryUnbond(ctx, &seq, msg.GetDecreaseAmount()); err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	// TODO: write object
	return &types.MsgDecreaseBondResponse{}, nil
}

func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.GetRealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	seq.OptedIn = false
	err = k.tryUnbond(ctx, &seq, seq.TokensCoin())
	if errorsmod.IsOf(types.ErrUnbondProposerOrSuccessor) {
		k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &seq.NoticePeriodTime,
			},
		}, nil
	}
	if err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	// TODO: write seq
	return &types.MsgUnbondResponse{}, nil
}
