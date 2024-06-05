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

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	demandOrder, err := m.ValidateOrderMutable(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// Check that the demand order fee is higher than the minimum fee
	minFee, _ := sdk.NewIntFromString(msg.MinFee)
	for _, coin := range demandOrder.Fee {
		if coin.Amount.LT(minFee) {
			return nil, types.ErrMinFeeNotMet
		}
	}
	// Check for blocked address
	if m.BankKeeper.BlockedAddr(demandOrder.GetRecipientBech32Address()) {
		return nil, types.ErrBlockedAddress
	}
	// Check that the fulfiller has enough balance to fulfill the order
	fulfillerAccount := m.GetAccount(ctx, msg.GetFulfillerBech32Address())
	if fulfillerAccount == nil {
		return nil, types.ErrFulfillerAddressDoesNotExist
	}
	// Send the funds from the fulfiller to the eibc packet original recipient
	err = m.BankKeeper.SendCoins(ctx, fulfillerAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), demandOrder.Price)
	if err != nil {
		logger.Error("Failed to send coins", "error", err)
		return nil, err
	}
	// Fulfill the order by updating the order status and underlying packet recipient
	err = m.Keeper.FulfillOrder(ctx, demandOrder, fulfillerAccount.GetAddress())

	return &types.MsgFulfillOrderResponse{}, err
}

// UpdateDemandOrder implements types.MsgServer.
func (m msgServer) UpdateDemandOrder(goCtx context.Context, msg *types.MsgUpdateDemandOrder) (*types.MsgUpdateDemandOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	// Check that the order exists in status PENDING
	demandOrder, err := m.ValidateOrderMutable(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// Check that the submitter is the expected recipient of the order
	submitter := msg.GetSubmitterAddr()
	if !submitter.Equals(demandOrder.GetRecipientBech32Address()) {
		return nil, types.ErrInvalidSubmitter
	}

	// TODO: check profitable

	// TODO: update the order (fee and price)
	demandOrder.Fee = msg.NewFee
	for _, coin := range demandOrder.Fee {
		if coin.Amount.LT(minFee) {
			return nil, types.ErrMinFeeNotMet
		}
	}

	err = m.SetDemandOrder(ctx, demandOrder)
	if err != nil {
		return err
	}
}

func (m msgServer) ValidateOrderMutable(ctx sdk.Context, orderId string) (*types.DemandOrder, error) {
	// Check that the order exists in status PENDING
	demandOrder, err := m.GetDemandOrder(ctx, commontypes.Status_PENDING, orderId)
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

	return demandOrder, nil
}
