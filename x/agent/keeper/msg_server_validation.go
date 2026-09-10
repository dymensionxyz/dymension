package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k msgServer) RequestValidation(goCtx context.Context, msg *types.MsgRequestValidation) (*types.MsgRequestValidationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	p, err := k.GetParams(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get params")
	}
	if uint64(len(msg.RequestUri)) > p.ValidationUriMaxBytes {
		return nil, errorsmod.Wrap(types.ErrInvalidValidationResponse, "request uri exceeds max bytes")
	}
	validator, found := k.GetAgent(ctx, msg.ValidatorId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.ValidatorId)
	}
	if !validator.Active {
		return nil, errorsmod.Wrap(types.ErrValidatorInactive, msg.ValidatorId)
	}
	if !k.IsAgentLive(ctx, msg.ValidatorId) {
		return nil, errorsmod.Wrap(types.ErrValidatorInactive, msg.ValidatorId)
	}
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if !agent.Active {
		return nil, errorsmod.Wrap(types.ErrValidatorInactive, msg.AgentId)
	}
	if !k.IsAgentLive(ctx, msg.AgentId) {
		return nil, errorsmod.Wrap(types.ErrValidatorInactive, msg.AgentId)
	}
	if msg.ValidatorId == msg.AgentId {
		return nil, types.ErrSelfValidation
	}
	if validator.Owner == agent.Owner {
		return nil, types.ErrSelfValidation
	}
	if msg.EvidenceSeq >= agent.ActionSeq {
		return nil, errorsmod.Wrapf(types.ErrInvalidEvidence, "evidence seq %d, agent action seq %d", msg.EvidenceSeq, agent.ActionSeq)
	}
	if has, err := k.validationRequests.Has(ctx, msg.RequestHash); err != nil {
		return nil, err
	} else if has {
		return nil, types.ErrValidationRequestExists
	}
	if !p.ValidationRequestFee.IsNil() && !p.ValidationRequestFee.IsZero() {
		coins := sdk.NewCoins(p.ValidationRequestFee)
		requester := sdk.MustAccAddressFromBech32(msg.Requester)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, requester, types.ModuleName, coins); err != nil {
			return nil, errorsmod.Wrap(types.ErrValidationFeePayment, err.Error())
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
			return nil, errorsmod.Wrap(types.ErrValidationFeePayment, err.Error())
		}
	}
	req := types.ValidationRequest{RequestHash: msg.RequestHash, Requester: msg.Requester, ValidatorId: msg.ValidatorId, AgentId: msg.AgentId, EvidenceSeq: msg.EvidenceSeq, RequestUri: msg.RequestUri, Height: ctx.BlockHeight(), Time: ctx.BlockTime()}
	if err := k.validationRequests.Set(ctx, msg.RequestHash, req); err != nil {
		return nil, err
	}
	if err := k.validationByAgent.Set(ctx, collections.Join(msg.AgentId, msg.RequestHash)); err != nil {
		return nil, err
	}
	if err := uevent.EmitTypedEvent(ctx, &types.EventRequestValidation{RequestHash: msg.RequestHash, ValidatorId: msg.ValidatorId, AgentId: msg.AgentId, EvidenceSeq: msg.EvidenceSeq}); err != nil {
		return nil, err
	}
	return &types.MsgRequestValidationResponse{}, nil
}

func (k msgServer) RespondValidation(goCtx context.Context, msg *types.MsgRespondValidation) (*types.MsgRespondValidationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	p, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if uint64(len(msg.Tag)) > p.ValidationTagMaxBytes || uint64(len(msg.ResponseUri)) > p.ValidationUriMaxBytes {
		return nil, errorsmod.Wrap(types.ErrInvalidValidationResponse, "tag or uri exceeds max bytes")
	}
	req, found := k.GetValidationRequest(ctx, msg.RequestHash)
	if !found {
		return nil, types.ErrValidationRequestNotFound
	}
	if p.ValidationMaxResponsesPerRequest != 0 && req.ResponseCount >= p.ValidationMaxResponsesPerRequest {
		return nil, types.ErrTooManyValidationResponses
	}
	validator, found := k.GetAgent(ctx, req.ValidatorId)
	if !found || !validator.Active {
		return nil, types.ErrValidatorInactive
	}
	if !k.IsAgentLive(ctx, req.ValidatorId) {
		return nil, types.ErrValidatorInactive
	}
	agent, found := k.GetAgent(ctx, req.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, req.AgentId)
	}
	if !agent.Active || !k.IsAgentLive(ctx, req.AgentId) {
		return nil, errorsmod.Wrap(types.ErrValidatorInactive, req.AgentId)
	}
	if req.ValidatorId == req.AgentId || validator.Owner == agent.Owner {
		return nil, types.ErrSelfValidation
	}
	if msg.Responder != validator.Owner {
		return nil, types.ErrUnauthorized
	}
	seq := req.ResponseCount
	resp := types.ValidationResponse{RequestHash: msg.RequestHash, Seq: seq, Response: msg.Response, ResponseUri: msg.ResponseUri, ResponseHash: msg.ResponseHash, Tag: msg.Tag, Height: ctx.BlockHeight(), Time: ctx.BlockTime()}
	if err := k.validationResponses.Set(ctx, collections.Join(msg.RequestHash, seq), resp); err != nil {
		return nil, err
	}
	req.ResponseCount++
	if err := k.validationRequests.Set(ctx, msg.RequestHash, req); err != nil {
		return nil, err
	}
	if err := uevent.EmitTypedEvent(ctx, &types.EventRespondValidation{RequestHash: msg.RequestHash, ValidatorId: req.ValidatorId, AgentId: req.AgentId, Response: msg.Response, Seq: seq}); err != nil {
		return nil, err
	}
	return &types.MsgRespondValidationResponse{Seq: seq}, nil
}
