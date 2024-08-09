package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// PlaceBuyOrder is message handler,
// handles creating an offer to buy a Dym-Name/Alias, performed by the buyer.
func (k msgServer) PlaceBuyOrder(goCtx context.Context, msg *dymnstypes.MsgPlaceBuyOrder) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)

	var resp *dymnstypes.MsgPlaceBuyOrderResponse
	var err error

	if msg.OrderType == dymnstypes.NameOrder {
		resp, err = k.placeBuyOrderTypeDymName(ctx, msg, params)
	} else if msg.OrderType == dymnstypes.AliasOrder {
		resp, err = k.placeBuyOrderTypeAlias(ctx, msg, params)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", msg.OrderType)
	}
	if err != nil {
		return nil, err
	}

	var minimumTxGasRequired sdk.Gas
	if msg.ContinueOfferId != "" {
		minimumTxGasRequired = dymnstypes.OpGasUpdateBuyOffer
	} else {
		minimumTxGasRequired = dymnstypes.OpGasPutBuyOffer
	}

	consumeMinimumGas(ctx, minimumTxGasRequired, "PlaceBuyOrder")

	return resp, nil
}

// placeBuyOrderTypeDymName handles the message handled by PlaceBuyOrder, type Dym-Name.
func (k msgServer) placeBuyOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	if !params.Misc.EnableTradingName {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Dym-Name is disabled")
	}

	existingOffer, err := k.validatePlaceBuyOrderTypeDymName(ctx, msg, params)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.BuyOffer
	var deposit sdk.Coin

	if existingOffer != nil {
		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetBuyOffer(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		deposit = msg.Offer

		offer = dymnstypes.BuyOffer{
			Id:         "", // will be auto-generated
			GoodsId:    msg.GoodsId,
			Type:       dymnstypes.NameOrder,
			Params:     msg.Params,
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

		err = k.AddReverseMappingGoodsIdToBuyOffer(ctx, msg.GoodsId, offer.Type, offer.Id)
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

	return &dymnstypes.MsgPlaceBuyOrderResponse{
		OfferId: offer.Id,
	}, nil
}

// validatePlaceBuyOrderTypeDymName handles validation for the message handled by PlaceBuyOrder, type Dym-Name.
func (k msgServer) validatePlaceBuyOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (existingOffer *dymnstypes.BuyOffer, err error) {
	dymName := k.GetDymNameWithExpirationCheck(ctx, msg.GoodsId)
	if dymName == nil {
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.GoodsId)
		return
	}
	if dymName.Owner == msg.Buyer {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "cannot buy own Dym-Name")
		return
	}

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
		if existingOffer.Type != msg.OrderType {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "order type mismatch with existing offer")
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

// placeBuyOrderTypeAlias handles the message handled by PlaceBuyOrder, type Alias.
func (k msgServer) placeBuyOrderTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	if !params.Misc.EnableTradingAlias {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Alias is disabled")
	}

	existingOffer, err := k.validatePlaceBuyOrderTypeAlias(ctx, msg, params)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.BuyOffer
	var deposit sdk.Coin

	if existingOffer != nil {
		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetBuyOffer(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		deposit = msg.Offer

		offer = dymnstypes.BuyOffer{
			Id:         "", // will be auto-generated
			GoodsId:    msg.GoodsId,
			Type:       dymnstypes.AliasOrder,
			Params:     msg.Params,
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

		err = k.AddReverseMappingGoodsIdToBuyOffer(ctx, msg.GoodsId, offer.Type, offer.Id)
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

	return &dymnstypes.MsgPlaceBuyOrderResponse{
		OfferId: offer.Id,
	}, nil
}

// validatePlaceBuyOrderTypeAlias handles validation for the message handled by PlaceBuyOrder, type Alias.
func (k msgServer) validatePlaceBuyOrderTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (existingOffer *dymnstypes.BuyOffer, err error) {
	destinationRollAppId := msg.Params[0]

	if !k.IsRollAppId(ctx, destinationRollAppId) {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "destination Roll-App does not exists: %s", destinationRollAppId)
		return
	}

	if !k.IsRollAppCreator(ctx, destinationRollAppId, msg.Buyer) {
		err = errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the RollApp: %s", destinationRollAppId)
		return
	}

	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, msg.GoodsId)
	if !found {
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not in-used: %s", msg.GoodsId)
		return
	}

	if destinationRollAppId == existingRollAppIdUsingAlias {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "destination Roll-App ID is the same as the source")
		return
	}

	if k.IsAliasPresentsInParamsAsAliasOrChainId(ctx, msg.GoodsId) {
		err = errorsmod.Wrapf(gerrc.ErrPermissionDenied,
			"prohibited to trade aliases which is reserved for chain-id or alias in module params: %s", msg.GoodsId,
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
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias mismatch with existing offer")
			return
		}
		if existingOffer.Type != msg.OrderType {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "order type mismatch with existing offer")
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
