package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

type msgServer struct {
	Keeper
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
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	panic("unimplemented")

}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
