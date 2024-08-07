package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelBuyOrder is message handler,
// handles canceling a Buy-Order, performed by the buyer who placed the offer.
func (k msgServer) CancelBuyOrder(goCtx context.Context, msg *dymnstypes.MsgCancelBuyOrder) (*dymnstypes.MsgCancelBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	offer := k.GetBuyOffer(ctx, msg.OfferId)
	if offer == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.OfferId)
	}

	var resp *dymnstypes.MsgCancelBuyOrderResponse
	var err error
	if offer.Type == dymnstypes.NameOrder || offer.Type == dymnstypes.AliasOrder {
		resp, err = k.processCancelBuyOrder(ctx, msg, *offer)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", offer.Type)
	}
	if err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseBuyOffer, "CancelBuyOrder")

	return resp, nil
}

// processCancelBuyOrder handles the message handled by CancelBuyOrder, type Dym-Name/Alias.
func (k msgServer) processCancelBuyOrder(
	ctx sdk.Context,
	msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOffer,
) (*dymnstypes.MsgCancelBuyOrderResponse, error) {
	if err := k.validateCancelOffer(ctx, msg, offer); err != nil {
		return nil, err
	}

	if err := k.RefundOffer(ctx, offer); err != nil {
		return nil, err
	}

	if err := k.removeBuyOffer(ctx, offer); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelBuyOrderResponse{}, nil
}

// validateCancelOffer handles validation for the message handled by CancelBuyOrder, type Dym-Name/Alias.
func (k msgServer) validateCancelOffer(_ sdk.Context, msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOffer) error {
	if offer.Buyer != msg.Buyer {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
	}

	return nil
}

// removeBuyOffer removes the Buy-Order from the store and the reverse mappings.
func (k msgServer) removeBuyOffer(ctx sdk.Context, offer dymnstypes.BuyOffer) error {
	k.DeleteBuyOffer(ctx, offer.Id)

	err := k.RemoveReverseMappingBuyerToBuyOffer(ctx, offer.Buyer, offer.Id)
	if err != nil {
		return err
	}

	err = k.RemoveReverseMappingGoodsIdToBuyOffer(ctx, offer.GoodsId, offer.Type, offer.Id)
	if err != nil {
		return err
	}

	return nil
}
