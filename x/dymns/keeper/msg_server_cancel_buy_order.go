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
	if offer.Type == dymnstypes.NameOrder {
		resp, err = k.processCancelBuyOrderTypeDymName(ctx, msg, *offer)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", offer.Type)
	}
	if err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseBuyOffer, "CancelBuyOrder")

	return resp, nil
}

// processCancelBuyOrderTypeDymName handles the message handled by CancelBuyOrder, type Dym-Name.
func (k msgServer) processCancelBuyOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOffer,
) (*dymnstypes.MsgCancelBuyOrderResponse, error) {
	if err := k.validateCancelOfferTypeDymName(ctx, msg, offer); err != nil {
		return nil, err
	}

	if err := k.RefundOffer(ctx, offer); err != nil {
		return nil, err
	}

	if err := k.removeBuyOfferTypeDymName(ctx, offer); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelBuyOrderResponse{}, nil
}

// validateCancelOfferTypeDymName handles validation for the message handled by CancelBuyOrder, type Dym-Name.
func (k msgServer) validateCancelOfferTypeDymName(_ sdk.Context, msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOffer) error {
	if offer.Buyer != msg.Buyer {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
	}

	return nil
}

// removeBuyOfferTypeDymName removes the Buy-Order from the store and the reverse mappings, type Dym-Name.
func (k msgServer) removeBuyOfferTypeDymName(ctx sdk.Context, offer dymnstypes.BuyOffer) error {
	k.DeleteBuyOffer(ctx, offer.Id)

	err := k.RemoveReverseMappingBuyerToBuyOffer(ctx, offer.Buyer, offer.Id)
	if err != nil {
		return err
	}

	err = k.RemoveReverseMappingDymNameToBuyOffer(ctx, offer.GoodsId, offer.Id)
	if err != nil {
		return err
	}

	return nil
}
