package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// TODO DymNS: buyer should be Roll-App owner

// PlaceBuyOrder is message handler,
// handles creating an offer to buy a Dym-Name/Alias, performed by the buyer.
func (k msgServer) PlaceBuyOrder(goCtx context.Context, msg *dymnstypes.MsgPlaceBuyOrder) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.OrderType != dymnstypes.MarketOrderType_MOT_DYM_NAME {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", msg.OrderType)
	}

	existingOffer, err := k.validateOffer(ctx, msg)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.BuyOffer
	var deposit sdk.Coin
	var minimumTxGasRequired sdk.Gas

	if existingOffer != nil {
		minimumTxGasRequired = dymnstypes.OpGasUpdateBuyOffer

		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetBuyOffer(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		minimumTxGasRequired = dymnstypes.OpGasPutBuyOffer

		deposit = msg.Offer

		offer = dymnstypes.BuyOffer{
			Id:         "", // will be auto-generated
			GoodsId:    msg.GoodsId,
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      msg.Buyer,
			OfferPrice: msg.Offer,
		}

		offer, err = k.InsertNewBuyOffer(ctx, offer)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingBuyerToBuyOfferRecord(ctx, msg.Buyer, offer.Id)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingDymNameToBuyOffer(ctx, msg.GoodsId, offer.Id)
		if err != nil {
			return nil, err
		}
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		sdk.MustAccAddressFromBech32(msg.Buyer),
		dymnstypes.ModuleName,
		sdk.NewCoins(deposit),
	); err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, minimumTxGasRequired, "PlaceBuyOrder")

	return &dymnstypes.MsgPlaceBuyOrderResponse{
		OfferId: offer.Id,
	}, nil
}

// validateOffer handles validation for the message handled by PlaceBuyOrder.
func (k msgServer) validateOffer(ctx sdk.Context, msg *dymnstypes.MsgPlaceBuyOrder) (existingOffer *dymnstypes.BuyOffer, err error) {
	err = msg.ValidateBasic()
	if err != nil {
		return
	}

	dymName := k.GetDymNameWithExpirationCheck(ctx, msg.GoodsId)
	if dymName == nil {
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.GoodsId)
		return
	}
	if dymName.Owner == msg.Buyer {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "cannot buy own Dym-Name")
		return
	}

	params := k.GetParams(ctx)
	if dymName.IsProhibitedTradingAt(ctx.BlockTime(), k.GetParams(ctx).Misc.ProhibitSellDuration) {
		err = errorsmod.Wrapf(gerrc.ErrFailedPrecondition,
			"duration before Dym-Name expiry, prohibited to trade: %s",
			params.Misc.ProhibitSellDuration,
		)
		return
	}

	if msg.Offer.Denom != params.Price.PriceDenom {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"invalid offer denomination, only accept %s", params.Price.PriceDenom,
		)
		return
	}
	if msg.Offer.Amount.LT(params.Price.MinOfferPrice) {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"offer price must be greater than or equal to %s", params.Price.MinOfferPrice.String(),
		)
		return
	}

	if msg.ContinueOfferId != "" {
		existingOffer = k.GetBuyOffer(ctx, msg.ContinueOfferId)
		if existingOffer == nil {
			err = errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.ContinueOfferId)
			return
		}
		if existingOffer.Buyer != msg.Buyer {
			err = errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
			return
		}
		if existingOffer.GoodsId != msg.GoodsId {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name mismatch with existing offer")
			return
		}
		if existingOffer.OfferPrice.Denom != msg.Offer.Denom {
			err = errorsmod.Wrapf(
				gerrc.ErrInvalidArgument,
				"offer denomination mismatch with existing offer: %s != %s", msg.Offer.Denom, existingOffer.OfferPrice.Denom,
			)
			return
		}
		if msg.Offer.IsLTE(existingOffer.OfferPrice) {
			err = errorsmod.Wrapf(
				gerrc.ErrInvalidArgument,
				"offer price must be greater than existing offer price %s", existingOffer.OfferPrice.String(),
			)
			return
		}
	}

	return
}
