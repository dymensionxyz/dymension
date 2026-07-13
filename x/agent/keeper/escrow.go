package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// GetEscrowBalance returns the agent's escrow balance; a missing ledger entry
// is an empty balance.
func (k Keeper) GetEscrowBalance(ctx sdk.Context, agentID string) sdk.Coins {
	e, err := k.escrows.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.NewCoins()
		}
		panic(err)
	}
	return e.Balance
}

// setEscrowBalance writes the agent's ledger entry, removing it when the
// balance is zero so the ledger only holds funded agents.
func (k Keeper) setEscrowBalance(ctx sdk.Context, agentID string, balance sdk.Coins) error {
	if balance.IsZero() {
		return k.escrows.Remove(ctx, agentID)
	}
	return k.escrows.Set(ctx, agentID, types.AgentEscrow{AgentId: agentID, Balance: balance})
}

func (k msgServer) FundAgentEscrow(goCtx context.Context, msg *types.MsgFundAgentEscrow) (*types.MsgFundAgentEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	if _, found := k.GetAgent(ctx, msg.AgentId); !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}

	funder := sdk.MustAccAddressFromBech32(msg.Funder)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, funder, types.ModuleName, msg.Amount); err != nil {
		return nil, errorsmod.Wrap(err, "send coins to module")
	}

	balance := k.GetEscrowBalance(ctx, msg.AgentId).Add(msg.Amount...)
	if err := k.setEscrowBalance(ctx, msg.AgentId, balance); err != nil {
		return nil, errorsmod.Wrap(err, "set escrow balance")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventFundAgentEscrow{
		AgentId: msg.AgentId,
		Funder:  msg.Funder,
		Amount:  msg.Amount,
	}); err != nil {
		return nil, err
	}

	return &types.MsgFundAgentEscrowResponse{}, nil
}

func (k msgServer) WithdrawAgentEscrow(goCtx context.Context, msg *types.MsgWithdrawAgentEscrow) (*types.MsgWithdrawAgentEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if agent.Owner != msg.Owner {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "not the agent owner")
	}

	balance, negative := k.GetEscrowBalance(ctx, msg.AgentId).SafeSub(msg.Amount...)
	if negative {
		return nil, errorsmod.Wrap(types.ErrInsufficientEscrow, msg.Amount.String())
	}
	if err := k.setEscrowBalance(ctx, msg.AgentId, balance); err != nil {
		return nil, errorsmod.Wrap(err, "set escrow balance")
	}

	owner := sdk.MustAccAddressFromBech32(msg.Owner)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, msg.Amount); err != nil {
		return nil, errorsmod.Wrap(err, "send coins to owner")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventWithdrawAgentEscrow{
		AgentId: msg.AgentId,
		Owner:   msg.Owner,
		Amount:  msg.Amount,
	}); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawAgentEscrowResponse{}, nil
}

func (k msgServer) UpdateAgentSpendPolicy(goCtx context.Context, msg *types.MsgUpdateAgentSpendPolicy) (*types.MsgUpdateAgentSpendPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrAgentNotFound, msg.AgentId)
	}
	if agent.Owner != msg.Owner {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "not the agent owner")
	}

	// The old window bookkeeping is meaningless under a new policy (the denom
	// or window length may have changed), so restart the window. This is not a
	// budget bypass: the spend policy is the owner's own guardrail over its own
	// escrow and is deliberately not timelocked.
	agent.SpendDenom = msg.SpendDenom
	agent.SpendLimitPerWindow = msg.SpendLimitPerWindow
	agent.SpendWindowBlocks = msg.SpendWindowBlocks
	agent.SpendWindowStartHeight = 0
	agent.SpendWindowSpent = math.ZeroInt()
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, errorsmod.Wrap(err, "set agent")
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventUpdateAgentSpendPolicy{
		AgentId:             msg.AgentId,
		SpendDenom:          msg.SpendDenom,
		SpendLimitPerWindow: msg.SpendLimitPerWindow,
		SpendWindowBlocks:   msg.SpendWindowBlocks,
	}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAgentSpendPolicyResponse{}, nil
}
