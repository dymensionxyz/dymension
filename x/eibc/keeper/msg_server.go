package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
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

func (m msgServer) FulfillOrder(goCtx context.Context, msg *types.MsgFulfillOrder) (*types.MsgFulfillOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()
	// Check that the msg is valid
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	// Check that the order exists in status PENDING
	demandOrder, err := m.GetDemandOrder(ctx, commontypes.Status_PENDING, msg.OrderId)
	if err != nil {
		return nil, err
	}
	// Check that the order is not fulfilled yet
	if demandOrder.IsFulfilled {
		return nil, types.ErrDemandAlreadyFulfilled
	}

	// Check the underlying packet is still relevant (i.e not expired, rejected, reverted)
	if demandOrder.TrackingPacketStatus != commontypes.Status_PENDING {
		return nil, types.ErrDemandOrderInactive
	}
	// Check for blocked address
	if m.bk.BlockedAddr(demandOrder.GetRecipientBech32Address()) {
		return nil, types.ErrBlockedAddress
	}
	// Check that the fulfiller has enough balance to fulfill the order
	fulfillerAccount := m.ak.GetAccount(ctx, msg.GetFulfillerBech32Address())
	if fulfillerAccount == nil {
		return nil, types.ErrFulfillerAddressDoesNotExist
	}

	// Send the funds from the fulfiller to the eibc packet original recipient
	err = m.bk.SendCoins(ctx, fulfillerAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), demandOrder.Price)
	if err != nil {
		logger.Error("Failed to send coins", "error", err)
		return nil, err
	}
	// Fulfill the order by updating the order status and underlying packet recipient
	err = m.Keeper.SetOrderFulfilled(ctx, demandOrder, fulfillerAccount.GetAddress())
	if err != nil {
		return nil, err
	}

	return &types.MsgFulfillOrderResponse{}, err
}
