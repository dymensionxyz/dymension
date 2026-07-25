package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k msgServer) SubmitFeedback(goCtx context.Context, msg *types.MsgSubmitFeedback) (*types.MsgSubmitFeedbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get params")
	}
	if uint64(len(msg.Tag1)) > params.FeedbackTagMaxBytes || uint64(len(msg.Tag2)) > params.FeedbackTagMaxBytes {
		return nil, errorsmod.Wrapf(types.ErrInvalidTag, "tag exceeds max bytes: max %d", params.FeedbackTagMaxBytes)
	}

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if msg.Client == agent.Owner {
		return nil, errorsmod.Wrap(types.ErrSelfFeedback, msg.Client)
	}
	// seq < action_seq suffices to prove the referenced action exists: the log
	// is append-only and entries are never deleted. This also blocks rating
	// agents with zero action history.
	if msg.EvidenceSeq >= agent.ActionSeq {
		return nil, errorsmod.Wrapf(types.ErrInvalidEvidence, "evidence seq %d, agent action seq %d", msg.EvidenceSeq, agent.ActionSeq)
	}

	// charge the anti-spam fee: send to module then burn (mirrors agent registration)
	if !params.FeedbackFee.IsNil() && !params.FeedbackFee.IsZero() {
		client := sdk.MustAccAddressFromBech32(msg.Client)
		coins := sdk.NewCoins(params.FeedbackFee)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, client, types.ModuleName, coins); err != nil {
			return nil, errorsmod.Wrap(types.ErrFeedbackFeePayment, err.Error())
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
			return nil, errorsmod.Wrap(types.ErrFeedbackFeePayment, err.Error())
		}
	}

	rep, found := k.GetReputation(ctx, msg.AgentId)
	if !found {
		rep = types.Reputation{AgentId: msg.AgentId}
	}
	if old, hadOld := k.GetFeedback(ctx, msg.AgentId, msg.Client); hadOld {
		rep.ScoreSum = rep.ScoreSum - uint64(old.Score) + uint64(msg.Score)
	} else {
		rep.Count++
		rep.ScoreSum += uint64(msg.Score)
	}

	fb := types.Feedback{
		AgentId:     msg.AgentId,
		Client:      msg.Client,
		Score:       msg.Score,
		Tag1:        msg.Tag1,
		Tag2:        msg.Tag2,
		EvidenceSeq: msg.EvidenceSeq,
		Height:      ctx.BlockHeight(),
		Time:        ctx.BlockTime(),
	}
	if err := k.SetFeedback(ctx, fb); err != nil {
		return nil, errorsmod.Wrap(err, "set feedback")
	}
	if err := k.SetReputation(ctx, rep); err != nil {
		return nil, errorsmod.Wrap(err, "set reputation")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventSubmitFeedback{
		AgentId: msg.AgentId,
		Client:  msg.Client,
		Score:   msg.Score,
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitFeedbackResponse{}, nil
}

func (k msgServer) RevokeFeedback(goCtx context.Context, msg *types.MsgRevokeFeedback) (*types.MsgRevokeFeedbackResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	old, found := k.GetFeedback(ctx, msg.AgentId, msg.Client)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrFeedbackNotFound, "agent %s client %s", msg.AgentId, msg.Client)
	}

	// The aggregate exists whenever a feedback exists, so Get cannot miss and
	// the subtractions cannot underflow.
	rep, err := k.reputation.Get(ctx, msg.AgentId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get reputation")
	}
	rep.Count--
	rep.ScoreSum -= uint64(old.Score)
	if rep.Count == 0 {
		if err := k.reputation.Remove(ctx, msg.AgentId); err != nil {
			return nil, errorsmod.Wrap(err, "remove reputation")
		}
	} else if err := k.SetReputation(ctx, rep); err != nil {
		return nil, errorsmod.Wrap(err, "set reputation")
	}

	if err := k.feedback.Remove(ctx, collections.Join(msg.AgentId, msg.Client)); err != nil {
		return nil, errorsmod.Wrap(err, "remove feedback")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventRevokeFeedback{
		AgentId: msg.AgentId,
		Client:  msg.Client,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRevokeFeedbackResponse{}, nil
}
