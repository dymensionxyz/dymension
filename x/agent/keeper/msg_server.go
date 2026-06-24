package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RegisterAgent(goCtx context.Context, msg *types.MsgRegisterAgent) (*types.MsgRegisterAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	exists, err := k.HasAgent(ctx, msg.AgentId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorsmod.Wrap(types.ErrAgentExists, msg.AgentId)
	}

	// charge the registration fee: send to module then burn (mirrors rollapp app registration)
	owner := sdk.MustAccAddressFromBech32(msg.Owner)
	fee := sdk.NewCoins(k.AgentRegistrationFee(ctx))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, fee); err != nil {
		return nil, errorsmod.Wrap(types.ErrRegistrationFeePayment, err.Error())
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, fee); err != nil {
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
		AgentId: agent.Id,
		Owner:   agent.Owner,
	}); err != nil {
		return nil, err
	}

	return &types.MsgRegisterAgentResponse{}, nil
}

func (k msgServer) DeactivateAgent(goCtx context.Context, msg *types.MsgDeactivateAgent) (*types.MsgDeactivateAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	agent, err := k.GetAgent(ctx, msg.AgentId)
	if errors.Is(err, collections.ErrNotFound) {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if err != nil {
		return nil, err
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
