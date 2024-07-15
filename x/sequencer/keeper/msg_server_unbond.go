package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//TODO: msg.ValidateBasic?

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

	// todo: wrap in IsUnbondRequested
	if seq.UnbondRequestHeight != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer has already requested to unbond",
		)
	}

	var (
		completionTime time.Time
		err            error
	)
	// sequencer required for a notice period before unbonding
	if seq.Proposer || seq.NextProposer {
		completionTime, err = k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &completionTime,
			},
		}, err
	} else {
		completionTime, err = k.setSequencerToUnbonding(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_UnbondingCompletionTime{
				UnbondingCompletionTime: &completionTime,
			},
		}, err
	}
}

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) (time.Time, error) {
	completionTime := ctx.BlockHeader().Time.Add(k.NoticePeriod(ctx))

	seq.UnbondRequestHeight = ctx.BlockHeight()
	seq.UnbondTime = completionTime

	k.UpdateSequencer(ctx, *seq, types.Bonded) // only bonded sequencers can have notice period

	k.SetNoticePeriodQueue(ctx, *seq)

	// TODO: emit notice period started event

	return completionTime, nil
}

func (k Keeper) setSequencerToUnbonding(ctx sdk.Context, seq *types.Sequencer) (time.Time, error) {
	oldStatus := seq.Status

	// set the status to unbonding
	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))

	// todo: wrap in seq.SetToUnbonding
	seq.Status = types.Unbonding
	seq.Proposer = false
	seq.NextProposer = false
	seq.UnbondTime = completionTime

	// don't overwrite the unbond request height in case notice period is already started
	if seq.UnbondRequestHeight == 0 {
		seq.UnbondRequestHeight = ctx.BlockHeight()
	}

	k.UpdateSequencer(ctx, *seq, oldStatus)
	k.SetUnbondingSequencerQueue(ctx, *seq)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbonding,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.SequencerAddress),
			sdk.NewAttribute(types.AttributeKeyBond, seq.Tokens.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.String()),
		),
	)

	return completionTime, nil
}
