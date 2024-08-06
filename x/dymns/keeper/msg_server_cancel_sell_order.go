package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelSellOrder is message handler,
// handles canceling Sell-Order, performed by the owner.
// This will stop the advertisement and remove the Dym-Name/Alias sale from the market.
// Can only be performed if no one has placed a bid on the goods.
func (k msgServer) CancelSellOrder(goCtx context.Context, msg *dymnstypes.MsgCancelSellOrder) (*dymnstypes.MsgCancelSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	var resp *dymnstypes.MsgCancelSellOrderResponse
	var err error
	if msg.OrderType == dymnstypes.NameOrder {
		resp, err = k.processCancelSellOrderTypeDymName(ctx, msg)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", msg.OrderType)
	}
	if err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseSellOrder, "CancelSellOrder")

	return resp, nil
}

// processCancelSellOrderTypeDymName handles the message handled by CancelSellOrder, type Dym-Name.
func (k msgServer) processCancelSellOrderTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) (*dymnstypes.MsgCancelSellOrderResponse, error) {
	if err := k.validateCancelSellOrderTypeDymName(ctx, msg); err != nil {
		return nil, err
	}

	k.DeleteSellOrder(ctx, msg.GoodsId, msg.OrderType)

	aSoe := k.GetActiveSellOrdersExpiration(ctx, msg.OrderType)
	aSoe.Remove(msg.GoodsId)
	if err := k.SetActiveSellOrdersExpiration(ctx, aSoe, msg.OrderType); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelSellOrderResponse{}, nil
}

// validateCancelSellOrderTypeDymName handles validation for the message handled by CancelSellOrder, type Dym-Name.
func (k msgServer) validateCancelSellOrderTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) error {
	dymName := k.GetDymName(ctx, msg.GoodsId)
	if dymName == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.GoodsId)
	}

	if dymName.Owner != msg.Owner {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	so := k.GetSellOrder(ctx, msg.GoodsId, msg.OrderType)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.GoodsId)
	}

	if so.HasExpiredAtCtx(ctx) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel an expired order")
	}

	if so.HighestBid != nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel once bid placed")
	}

	return nil
}
