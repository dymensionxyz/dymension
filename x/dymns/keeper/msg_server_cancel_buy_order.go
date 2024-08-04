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

	offer, err := k.validateCancelOffer(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err := k.RefundOffer(ctx, *offer); err != nil {
		return nil, err
	}

	if err := k.removeBuyOffer(ctx, *offer); err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseBuyOffer, "CancelBuyOrder")

	return &dymnstypes.MsgCancelBuyOrderResponse{}, nil
}

// validateCancelOffer handles validation for the message handled by CancelBuyOrder.
func (k msgServer) validateCancelOffer(ctx sdk.Context, msg *dymnstypes.MsgCancelBuyOrder) (*dymnstypes.BuyOffer, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	offer := k.GetBuyOffer(ctx, msg.OfferId)
	if offer == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.OfferId)
	}

	if offer.Buyer != msg.Buyer {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
	}

	return offer, nil
}

// removeBuyOffer removes the Buy-Order from the store and the reverse mappings.
func (k msgServer) removeBuyOffer(ctx sdk.Context, offer dymnstypes.BuyOffer) error {
	k.DeleteBuyOffer(ctx, offer.Id)

	err := k.RemoveReverseMappingBuyerToBuyOffer(ctx, offer.Buyer, offer.Id)
	if err != nil {
		return err
	}

	err = k.RemoveReverseMappingDymNameToBuyOffer(ctx, offer.Name, offer.Id)
	if err != nil {
		return err
	}

	return nil
}
