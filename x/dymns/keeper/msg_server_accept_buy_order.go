package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// AcceptBuyOrder is message handler,
// handles accepting a Buy-Offer or raising the amount for negotiation,
// performed by the owner of the goods.
func (k msgServer) AcceptBuyOrder(goCtx context.Context, msg *dymnstypes.MsgAcceptBuyOrder) (*dymnstypes.MsgAcceptBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	bo := k.GetBuyOffer(ctx, msg.OfferId)
	if bo == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order: %s", msg.OfferId)
	}

	params := k.GetParams(ctx)

	var resp *dymnstypes.MsgAcceptBuyOrderResponse
	var err error

	if bo.Type == dymnstypes.NameOrder {
		resp, err = k.processAcceptBuyOrderTypeDymName(ctx, msg, *bo, params)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", bo.Type)
	}
	if err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasUpdateBuyOffer, "AcceptBuyOrder")

	return resp, nil
}

// processAcceptBuyOrderTypeDymName handles the message handled by AcceptBuyOrder, type Dym-Name.
func (k msgServer) processAcceptBuyOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, offer dymnstypes.BuyOffer, params dymnstypes.Params,
) (*dymnstypes.MsgAcceptBuyOrderResponse, error) {
	if !params.Misc.EnableTradingName {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Dym-Name is disabled")
	}

	dymName, err := k.validateAcceptBuyOrderTypeDymName(ctx, msg, offer, params)
	if err != nil {
		return nil, err
	}

	var accepted bool

	if msg.MinAccept.IsLT(offer.OfferPrice) {
		// this was checked earlier so this won't happen,
		// but I keep this here to easier to understand of all-cases of comparison
		panic("min-accept is less than offer price")
	} else if msg.MinAccept.IsEqual(offer.OfferPrice) {
		accepted = true

		// take the offer
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			dymnstypes.ModuleName,
			sdk.MustAccAddressFromBech32(dymName.Owner),
			sdk.Coins{offer.OfferPrice},
		); err != nil {
			return nil, err
		}

		if err := k.removeBuyOffer(ctx, offer); err != nil {
			return nil, err
		}

		if err := k.transferOwnership(ctx, *dymName, offer.Buyer); err != nil {
			return nil, err
		}
	} else {
		accepted = false

		offer.CounterpartyOfferPrice = &msg.MinAccept
		if err := k.SetBuyOffer(ctx, offer); err != nil {
			return nil, err
		}
	}

	return &dymnstypes.MsgAcceptBuyOrderResponse{
		Accepted: accepted,
	}, nil
}

// validateAcceptBuyOrderTypeDymName handles validation for the message handled by AcceptBuyOrder, type Dym-Name.
func (k msgServer) validateAcceptBuyOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, bo dymnstypes.BuyOffer, params dymnstypes.Params,
) (*dymnstypes.DymName, error) {
	dymName := k.GetDymNameWithExpirationCheck(ctx, bo.GoodsId)
	if dymName == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", bo.GoodsId)
	}

	if dymName.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	if dymName.IsProhibitedTradingAt(ctx.BlockTime(), params.Misc.ProhibitSellDuration) {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition,
			"duration before Dym-Name expiry, prohibited to sell: %s",
			params.Misc.ProhibitSellDuration,
		)
	}

	if bo.Buyer == msg.Owner {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "cannot accept own offer")
	}

	if msg.MinAccept.Denom != bo.OfferPrice.Denom {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"denom must be the same as the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	if msg.MinAccept.IsLT(bo.OfferPrice) {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"amount must be greater than or equals to the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	return dymName, nil
}
