package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelOfferBuyName is message handler,
// handles canceling an Offer-ToBuy, performed by the buyer who placed the offer.
func (k msgServer) CancelOfferBuyName(goCtx context.Context, msg *dymnstypes.MsgCancelOfferBuyName) (*dymnstypes.MsgCancelOfferBuyNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	offer, err := k.validateCancelOffer(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err := k.RefundOffer(ctx, *offer); err != nil {
		return nil, err
	}

	if err := k.removeOfferToBuy(ctx, *offer); err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseOffer, "CancelOfferBuyName")

	return &dymnstypes.MsgCancelOfferBuyNameResponse{}, nil
}

// validateCancelOffer handles validation for the message handled by CancelOfferBuyName.
func (k msgServer) validateCancelOffer(ctx sdk.Context, msg *dymnstypes.MsgCancelOfferBuyName) (*dymnstypes.OfferToBuy, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	offer := k.GetOfferToBuy(ctx, msg.OfferId)
	if offer == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Offer-To-Buy: %s", msg.OfferId)
	}

	if offer.Buyer != msg.Buyer {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
	}

	return offer, nil
}

// removeOfferToBuy removes the Offer-To-Buy from the store and the reverse mappings.
func (k msgServer) removeOfferToBuy(ctx sdk.Context, offer dymnstypes.OfferToBuy) error {
	k.DeleteOfferToBuy(ctx, offer.Id)

	err := k.RemoveReverseMappingBuyerToOfferToBuy(ctx, offer.Buyer, offer.Id)
	if err != nil {
		return err
	}

	err = k.RemoveReverseMappingDymNameToOfferToBuy(ctx, offer.Name, offer.Id)
	if err != nil {
		return err
	}

	return nil
}
