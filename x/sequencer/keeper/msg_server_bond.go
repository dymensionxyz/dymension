package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.RealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if err := validBondDenom(msg.AddAmount); err != nil {
		return nil, err
	}

	// charge the user and modify the sequencer object
	if err := k.sendToModule(ctx, &seq, msg.AddAmount); err != nil {
		return nil, err
	}
	k.SetSequencer(ctx, seq)

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

	seq, err := k.RealSequencer(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	if err := k.TryUnbond(ctx, &seq, msg.GetDecreaseAmount()); err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	k.SetSequencer(ctx, seq)

	return &types.MsgDecreaseBondResponse{}, nil
}

func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	// not allowed to unbond immediately, need to serve a notice to allow the rollapp community to organise
	// Also, if they already requested to unbond, we don't want to start another notice period, regardless
	// of if their notice already elapsed or not.
	if k.AwaitingLastProposerBlock(ctx, seq.RollappId) && (k.IsProposer(ctx, seq) || k.IsSuccessor(ctx, seq)) {
		return nil, gerrc.ErrFailedPrecondition.Wrap("cannot unbond while rotation in progress")
	}

	// ensures they will not get chosen as their own successor!
	if err := seq.SetOptedIn(ctx, false); err != nil {
		return nil, err
	}

	// now we know they are proposer
	// avoid starting another notice unnecessarily
	if k.IsProposer(ctx, seq) {
		if !k.rollappKeeper.ForkLatestAllowed(ctx, seq.RollappId) {
			return nil, gerrc.ErrFailedPrecondition.Wrap("rotation could cause fork before genesis transfer")
		}
		if seq.NoticeInProgress(ctx.BlockTime()) {
			return nil, gerrc.ErrFailedPrecondition.Wrap("notice period in progress")
		}

		k.StartNoticePeriod(ctx, &seq)
		k.SetSequencer(ctx, seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &seq.NoticePeriodTime,
			},
		}, nil

	}

	err = k.TryUnbond(ctx, &seq, seq.TokensCoin())
	if err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}
	k.SetSequencer(ctx, seq)

	return &types.MsgUnbondResponse{}, nil
}
