package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type msgServer struct {
	Keeper
}

// CreatePlan implements types.MsgServer.
func (m msgServer) CreatePlan(goCtx context.Context, req *types.MsgCreatePlan) (*types.MsgCreatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, found := m.Keeper.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	planId, err := m.Keeper.CreatePlan(ctx, req.AllocatedAmount, req.StartTime, req.EndTime, rollapp)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePlanResponse{
		PlanId: planId,
	}, nil

}

// Buy implements types.MsgServer.
func (m msgServer) Buy(context.Context, *types.MsgBuy) (*types.MsgBuyResponse, error) {
	panic("unimplemented")
}

// Claim implements types.MsgServer.
func (m msgServer) Claim(context.Context, *types.MsgClaim) (*types.MsgClaimResponse, error) {
	panic("unimplemented")
}

// Sell implements types.MsgServer.
func (m msgServer) Sell(context.Context, *types.MsgSell) (*types.MsgSellResponse, error) {
	panic("unimplemented")
}

// UpdateParams implements types.MsgServer.
func (m msgServer) UpdateParams(context.Context, *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
