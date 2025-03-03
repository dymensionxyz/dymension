package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// TradingEnable implements types.MsgServer.
func (m msgServer) EnableTrading(ctx context.Context, req *types.MsgEnableTrading) (*types.MsgEnableTradingResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.EnableTrading(sdk.UnwrapSDKContext(ctx), req.PlanId, owner)
	if err != nil {
		return nil, err
	}

	return &types.MsgEnableTradingResponse{}, nil
}

// Buy implements types.MsgServer.
func (m msgServer) Buy(ctx context.Context, req *types.MsgBuy) (*types.MsgBuyResponse, error) {
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.Buy(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Amount, req.MaxCostAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{}, nil
}

// BuyExactSpend implements types.MsgServer.
func (m msgServer) BuyExactSpend(ctx context.Context, req *types.MsgBuyExactSpend) (*types.MsgBuyResponse, error) {
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.BuyExactSpend(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Spend, req.MinOutTokensAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{}, nil
}

// Sell implements types.MsgServer.
func (m msgServer) Sell(ctx context.Context, req *types.MsgSell) (*types.MsgSellResponse, error) {
	seller, err := sdk.AccAddressFromBech32(req.Seller)
	if err != nil {
		return nil, err
	}
	err = m.Keeper.Sell(sdk.UnwrapSDKContext(ctx), req.PlanId, seller, req.Amount, req.MinIncomeAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellResponse{}, nil
}
