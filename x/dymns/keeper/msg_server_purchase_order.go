package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// PurchaseOrder is message handler,
// handles purchasing a Dym-Name/Alias from a Sell-Order, performed by the buyer.
func (k msgServer) PurchaseOrder(goCtx context.Context, msg *dymnstypes.MsgPurchaseOrder) (*dymnstypes.MsgPurchaseOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.OrderType != dymnstypes.NameOrder {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", msg.OrderType)
	}

	dymName, so, err := k.validatePurchase(ctx, msg)
	if err != nil {
		return nil, err
	}

	if so.HighestBid != nil {
		// refund previous bidder
		if err := k.RefundBid(ctx, *so.HighestBid, so.Type); err != nil {
			return nil, err
		}
	}

	// deduct offer price from buyer's account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sdk.MustAccAddressFromBech32(msg.Buyer),
		dymnstypes.ModuleName,
		sdk.Coins{msg.Offer},
	); err != nil {
		return nil, err
	}

	// record new highest bid
	so.HighestBid = &dymnstypes.SellOrderBid{
		Bidder: msg.Buyer,
		Price:  msg.Offer,
		Params: msg.Params,
	}

	// after highest bid updated, update SO to store to reflect the new state
	if err := k.SetSellOrder(ctx, *so); err != nil {
		return nil, err
	}

	// try to complete the purchase

	if so.HasFinishedAtCtx(ctx) {
		if err := k.CompleteDymNameSellOrder(ctx, dymName.Name); err != nil {
			return nil, err
		}
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasPlaceBidOnSellOrder, "PurchaseOrder")

	return &dymnstypes.MsgPurchaseOrderResponse{}, nil
}

// validatePurchase handles validation for the message handled by PurchaseOrder.
func (k msgServer) validatePurchase(ctx sdk.Context, msg *dymnstypes.MsgPurchaseOrder) (*dymnstypes.DymName, *dymnstypes.SellOrder, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, nil, err
	}

	dymName := k.GetDymName(ctx, msg.GoodsId)
	if dymName == nil {
		return nil, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.GoodsId)
	}

	if dymName.Owner == msg.Buyer {
		return nil, nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "cannot purchase your own dym name")
	}

	so := k.GetSellOrder(ctx, msg.GoodsId, msg.OrderType)
	if so == nil {
		return nil, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.GoodsId)
	}

	if so.HasExpiredAtCtx(ctx) {
		return nil, nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot purchase an expired order")
	}

	if so.HasFinishedAtCtx(ctx) {
		return nil, nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot purchase a completed order")
	}

	if msg.Offer.Denom != so.MinPrice.Denom {
		return nil, nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"offer denom does not match the order denom: %s != %s",
			msg.Offer.Denom, so.MinPrice.Denom,
		)
	}

	if msg.Offer.IsLT(so.MinPrice) {
		return nil, nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer is lower than minimum price")
	}

	if so.HasSetSellPrice() {
		if !msg.Offer.IsLTE(*so.SellPrice) { // overpaid protection
			return nil, nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer is higher than sell price")
		}
	}

	if so.HighestBid != nil {
		if msg.Offer.IsLTE(so.HighestBid.Price) {
			return nil, nil, errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"new offer must be higher than current highest bid",
			)
		}
	}

	return dymName, so, nil
}
