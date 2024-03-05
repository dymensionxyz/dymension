package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	completionTime, err := k.setSequencerToUnbonding(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	return &types.MsgUnbondResponse{
		CompletionTime: completionTime,
	}, nil
}

func (k Keeper) setSequencerToUnbonding(ctx sdk.Context, seqAddr string) (time.Time, error) {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return time.Time{}, types.ErrUnknownSequencer
	}

	if !seq.IsBonded() {
		return time.Time{}, sdkerrors.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not bonded: got %s",
			seq.Status.String(),
		)
	}
	oldStatus := seq.Status
	wasPropser := seq.Proposer

	// set the status to unbonding
	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	seq.Status = types.Unbonding
	seq.Proposer = false
	seq.UnbondingHeight = ctx.BlockHeight()
	seq.UnbondTime = completionTime

	k.UpdateSequencer(ctx, seq, oldStatus)
	k.setUnbondingSequencerQueue(ctx, seq)

	if wasPropser {
		k.RotateProposer(ctx, seq.RollappId)
	}

	//TODO: emit an event

	return completionTime, nil
}
