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

	if msg.AssetType == dymnstypes.TypeName {
		resp, err = k.placeBuyOrderWithAssetTypeDymName(ctx, msg, params)
	} else if msg.AssetType == dymnstypes.TypeAlias {
		resp, err = k.placeBuyOrderWithAssetTypeAlias(ctx, msg, params)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", msg.AssetType)
	}
	if err != nil {
		return nil, err
	}

	var minimumTxGasRequired sdk.Gas
	if msg.ContinueOrderId != "" {
		minimumTxGasRequired = dymnstypes.OpGasUpdateBuyOrder
	} else {
		minimumTxGasRequired = dymnstypes.OpGasPutBuyOrder
	}

	consumeMinimumGas(ctx, minimumTxGasRequired, "PlaceBuyOrder")

	return resp, nil
}

// placeBuyOrderWithAssetTypeDymName handles the message handled by PlaceBuyOrder, type Dym-Name.
func (k msgServer) placeBuyOrderWithAssetTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	if !params.Misc.EnableTradingName {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Dym-Name is disabled")
	}

	existingOffer, err := k.validatePlaceBuyOrderWithAssetTypeDymName(ctx, msg, params)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.BuyOrder
	var deposit sdk.Coin

	if existingOffer != nil {
		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetBuyOrder(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		deposit = msg.Offer

		offer = dymnstypes.BuyOrder{
			Id:         "", // will be auto-generated
			AssetId:    msg.AssetId,
			AssetType:  dymnstypes.TypeName,
			Params:     msg.Params,
			Buyer:      msg.Buyer,
			OfferPrice: msg.Offer,
		}

		offer, err = k.InsertNewBuyOrder(ctx, offer)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingBuyerToBuyOrderRecord(ctx, msg.Buyer, offer.Id)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingAssetIdToBuyOrder(ctx, msg.AssetId, offer.AssetType, offer.Id)
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
		OrderId: offer.Id,
	}, nil
}

// validatePlaceBuyOrderWithAssetTypeDymName handles validation for the message handled by PlaceBuyOrder, type Dym-Name.
func (k msgServer) validatePlaceBuyOrderWithAssetTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (existingOffer *dymnstypes.BuyOrder, err error) {
	dymName := k.GetDymNameWithExpirationCheck(ctx, msg.AssetId)
	if dymName == nil {
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.AssetId)
		return
	}
	if dymName.Owner == msg.Buyer {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "cannot buy own Dym-Name")
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

	if msg.ContinueOrderId != "" {
		existingOffer = k.GetBuyOrder(ctx, msg.ContinueOrderId)
		if existingOffer == nil {
			err = errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.ContinueOrderId)
			return
		}
		if existingOffer.Buyer != msg.Buyer {
			err = errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
			return
		}
		if existingOffer.AssetId != msg.AssetId {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name mismatch with existing offer")
			return
		}
		if existingOffer.AssetType != msg.AssetType {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "asset type mismatch with existing offer")
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

// placeBuyOrderWithAssetTypeAlias handles the message handled by PlaceBuyOrder, type Alias.
func (k msgServer) placeBuyOrderWithAssetTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (*dymnstypes.MsgPlaceBuyOrderResponse, error) {
	if !params.Misc.EnableTradingAlias {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Alias is disabled")
	}

	existingOffer, err := k.validatePlaceBuyOrderWithAssetTypeAlias(ctx, msg, params)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.BuyOrder
	var deposit sdk.Coin

	if existingOffer != nil {
		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetBuyOrder(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		deposit = msg.Offer

		offer = dymnstypes.BuyOrder{
			Id:         "", // will be auto-generated
			AssetId:    msg.AssetId,
			AssetType:  dymnstypes.TypeAlias,
			Params:     msg.Params,
			Buyer:      msg.Buyer,
			OfferPrice: msg.Offer,
		}

		offer, err = k.InsertNewBuyOrder(ctx, offer)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingBuyerToBuyOrderRecord(ctx, msg.Buyer, offer.Id)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingAssetIdToBuyOrder(ctx, msg.AssetId, offer.AssetType, offer.Id)
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
		OrderId: offer.Id,
	}, nil
}

// validatePlaceBuyOrderWithAssetTypeAlias handles validation for the message handled by PlaceBuyOrder, type Alias.
func (k msgServer) validatePlaceBuyOrderWithAssetTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceBuyOrder, params dymnstypes.Params,
) (existingOffer *dymnstypes.BuyOrder, err error) {
	destinationRollAppId := msg.Params[0]

	if !k.IsRollAppId(ctx, destinationRollAppId) {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "destination Roll-App does not exists: %s", destinationRollAppId)
		return
	}

	if !k.IsRollAppCreator(ctx, destinationRollAppId, msg.Buyer) {
		err = errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the RollApp: %s", destinationRollAppId)
		return
	}

	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, msg.AssetId)
	if !found {
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not in-used: %s", msg.AssetId)
		return
	}

	if destinationRollAppId == existingRollAppIdUsingAlias {
		err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "destination Roll-App ID is the same as the source")
		return
	}

	if k.IsAliasPresentsInParamsAsAliasOrChainId(ctx, msg.AssetId) {
		err = errorsmod.Wrapf(gerrc.ErrPermissionDenied,
			"prohibited to trade aliases which is reserved for chain-id or alias in module params: %s", msg.AssetId,
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

	if msg.ContinueOrderId != "" {
		existingOffer = k.GetBuyOrder(ctx, msg.ContinueOrderId)
		if existingOffer == nil {
			err = errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.ContinueOrderId)
			return
		}
		if existingOffer.Buyer != msg.Buyer {
			err = errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
			return
		}
		if existingOffer.AssetId != msg.AssetId {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias mismatch with existing offer")
			return
		}
		if existingOffer.AssetType != msg.AssetType {
			err = errorsmod.Wrap(gerrc.ErrInvalidArgument, "asset type mismatch with existing offer")
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
