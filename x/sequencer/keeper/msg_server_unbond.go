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

	if seq.UnbondRequestHeight != 0 {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer has already requested to unbond",
		)
	}
	seq.UnbondRequestHeight = ctx.BlockHeight()

	// check if sequencer required for a notice period before unbonding
	if seq.IsNoticePeriodRequired() {
		completionTime, err := k.startNoticePeriodForSequencer(ctx, &seq)
		return &types.MsgUnbondResponse{
			CompletionTime: &types.MsgUnbondResponse_NoticePeriodCompletionTime{
				NoticePeriodCompletionTime: &completionTime,
			},
		}, err
	}

	// otherwise, start unbonding
	completionTime, err := k.setSequencerToUnbonding(ctx, &seq)
	return &types.MsgUnbondResponse{
		CompletionTime: &types.MsgUnbondResponse_UnbondingCompletionTime{
			UnbondingCompletionTime: &completionTime,
		},
	}, err
}

func (k Keeper) startNoticePeriodForSequencer(ctx sdk.Context, seq *types.Sequencer) (time.Time, error) {
	completionTime := ctx.BlockHeader().Time.Add(k.NoticePeriod(ctx))
	seq.UnbondTime = completionTime

	k.UpdateSequencer(ctx, *seq, types.Bonded) // only bonded sequencers can have notice period
	k.SetNoticePeriodQueue(ctx, *seq)

	nextSeq := k.ExpectedNextProposer(ctx, seq.RollappId)
	if nextSeq.SequencerAddress == "" {
		k.Logger(ctx).Info("rollapp will be left with no proposer after notice period", "rollappId", seq.RollappId, "sequencer", seq.SequencerAddress)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNoticePeriodStarted,
			sdk.NewAttribute(types.AttributeKeySequencer, seq.SequencerAddress),
			sdk.NewAttribute(types.AttributeKeyNextProposer, nextSeq.SequencerAddress),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.String()),
		),
	)

	return completionTime, nil
}

// setSequencerToUnbonding sets the sequencer to unbonding status
// can be called after notice period or directly if notice period is not required
func (k Keeper) setSequencerToUnbonding(ctx sdk.Context, seq *types.Sequencer) (time.Time, error) {
	// set the status to unbonding
	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	seq.UnbondTime = completionTime

	seq.Status = types.Unbonding
	seq.Proposer = false
	seq.NextProposer = false

	k.UpdateSequencer(ctx, *seq, types.Bonded) // only bonded sequencers can start unbonding
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
