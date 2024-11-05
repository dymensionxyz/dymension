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
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	if err := k.validBondDenom(ctx, msg.AddAmount); err != nil {
		return nil, err
	}

	if err := k.sendToModule(ctx, &seq, msg.AddAmount); err != nil {
		return nil, err
	}

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
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	if err := k.TryUnbond(ctx, &seq, msg.GetDecreaseAmount()); err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}

	return &types.MsgDecreaseBondResponse{}, nil
}

func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	// ensures they will not get chosen as their own successor!
	if err := seq.SetOptedIn(ctx, false); err != nil {
		return nil, err
	}
	err = k.TryUnbond(ctx, &seq, seq.TokensCoin())
	if errorsmod.IsOf(err, types.ErrUnbondProposerOrSuccessor) {
		// not allowed to unbond immediately, need to serve a notice to allow the rollapp community to organise
		// Also, if they already requested to unbond, we don't want to start another notice period, regardless
		// of if their notice already elapsed or not.
		if k.IsSuccessor(ctx, seq) {
			return nil, gerrc.ErrFailedPrecondition.Wrap("successor cannot unbond or start notice")
		}
		// now we know they are proposer
		// avoid starting another notice unnecessarily
		if !k.RotationInProgress(ctx, seq.RollappId) {
			k.StartNoticePeriod(ctx, &seq)
		}
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &seq.NoticePeriodTime,
			},
		}, nil
	}
	if err != nil {
		return nil, errorsmod.Wrap(err, "try unbond")
	}

	return &types.MsgUnbondResponse{}, nil
}
