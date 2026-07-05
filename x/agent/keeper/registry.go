package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k msgServer) RegisterAgent(goCtx context.Context, msg *types.MsgRegisterAgent) (*types.MsgRegisterAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, found := k.GetAgent(ctx, msg.AgentId); found {
		return nil, errorsmod.Wrap(types.ErrAgentExists, msg.AgentId)
	}

	fp, err := types.PolicyFingerprint(msg.Policy)
	if err != nil {
		return nil, errorsmod.Wrap(err, "policy fingerprint")
	}
	revoked, err := k.IsPolicyRevoked(ctx, fp)
	if err != nil {
		return nil, errorsmod.Wrap(err, "is policy revoked")
	}
	if revoked {
		return nil, gerrc.ErrFailedPrecondition.Wrapf("policy revoked: %s", fp)
	}

	// charge the registration fee: send to module then burn (mirrors rollapp app registration)
	fee, err := k.AgentRegistrationFee(ctx)
	if err != nil {
		return nil, err
	}
	owner := sdk.MustAccAddressFromBech32(msg.Owner)
	coins := sdk.NewCoins(fee)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, coins); err != nil {
		return nil, errorsmod.Wrap(types.ErrRegistrationFeePayment, err.Error())
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, errorsmod.Wrap(types.ErrRegistrationFeePayment, err.Error())
	}

	agent := types.Agent{
		Id:        msg.AgentId,
		Owner:     msg.Owner,
		Policy:    msg.Policy,
		Active:    true,
		ActionSeq: 0,
	}
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventRegisterAgent{
		AgentId:     agent.Id,
		Owner:       agent.Owner,
		Fingerprint: fp,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAgentResponse{}, nil
}

func (k msgServer) DeactivateAgent(goCtx context.Context, msg *types.MsgDeactivateAgent) (*types.MsgDeactivateAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if agent.Owner != msg.Owner {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "not the agent owner")
	}

	agent.Active = false
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventDeactivateAgent{
		AgentId: agent.Id,
		Owner:   agent.Owner,
	}); err != nil {
		return nil, err
	}

	return &types.MsgDeactivateAgentResponse{}, nil
}
