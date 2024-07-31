package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (k msgServer) OfferBuyName(goCtx context.Context, msg *dymnstypes.MsgOfferBuyName) (*dymnstypes.MsgOfferBuyNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existingOffer, err := k.validateOffer(ctx, msg)
	if err != nil {
		return nil, err
	}

	var offer dymnstypes.OfferToBuy
	var deposit sdk.Coin
	var minimumTxGasRequired sdk.Gas

	if existingOffer != nil {
		minimumTxGasRequired = dymnstypes.OpGasUpdateOffer

		deposit = msg.Offer.Sub(existingOffer.OfferPrice)

		offer = *existingOffer
		offer.OfferPrice = msg.Offer

		if err := k.SetOfferToBuy(ctx, offer); err != nil {
			return nil, err
		}
	} else {
		minimumTxGasRequired = dymnstypes.OpGasPutOffer

		deposit = msg.Offer

		offer = dymnstypes.OfferToBuy{
			Id:         "", // will be auto-generated
			Name:       msg.Name,
			Buyer:      msg.Buyer,
			OfferPrice: msg.Offer,
		}

		offer, err = k.InsertOfferToBuy(ctx, offer)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingBuyerToOfferToBuyRecord(ctx, msg.Buyer, offer.Id)
		if err != nil {
			return nil, err
		}

		err = k.AddReverseMappingDymNameToOfferToBuy(ctx, msg.Name, offer.Id)
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

	consumeMinimumGas(ctx, minimumTxGasRequired, "OfferBuyName")

	return &dymnstypes.MsgOfferBuyNameResponse{
		OfferId: offer.Id,
	}, nil
}

func (k msgServer) validateOffer(ctx sdk.Context, msg *dymnstypes.MsgOfferBuyName) (existingOffer *dymnstypes.OfferToBuy, err error) {
	err = msg.ValidateBasic()
	if err != nil {
		return
	}

	dymName := k.GetDymNameWithExpirationCheck(ctx, msg.Name, ctx.BlockTime().Unix())
	if dymName == nil {
		err = dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
		return
	}
	if dymName.Owner == msg.Buyer {
		err = sdkerrors.ErrInvalidRequest.Wrap("cannot buy own Dym-Name")
		return
	}

	params := k.GetParams(ctx)

	if msg.Offer.Denom != params.Price.PriceDenom {
		err = sdkerrors.ErrInvalidRequest.Wrapf(
			"invalid offer denomination, only accept %s", params.Price.PriceDenom,
		)
		return
	}
	if msg.Offer.Amount.LT(params.Price.MinOfferPrice) {
		err = sdkerrors.ErrInvalidRequest.Wrapf(
			"offer price must be greater than or equal to %s", params.Price.MinOfferPrice.String(),
		)
		return
	}

	if msg.ContinueOfferId != "" {
		existingOffer = k.GetOfferToBuy(ctx, msg.ContinueOfferId)
		if existingOffer == nil {
			err = dymnstypes.ErrOfferToBuyNotFound.Wrap(msg.ContinueOfferId)
			return
		}
		if existingOffer.Buyer != msg.Buyer {
			err = sdkerrors.ErrUnauthorized.Wrap("not the owner of the offer")
			return
		}
		if existingOffer.Name != msg.Name {
			err = sdkerrors.ErrInvalidRequest.Wrap("Dym-Name mismatch with existing offer")
			return
		}
		if existingOffer.OfferPrice.Denom != msg.Offer.Denom {
			err = sdkerrors.ErrInvalidRequest.Wrap("offer denomination mismatch with existing offer")
			return
		}
		if msg.Offer.IsLTE(existingOffer.OfferPrice) {
			err = sdkerrors.ErrInvalidRequest.Wrapf(
				"offer price must be greater than existing offer price %s",
				existingOffer.OfferPrice.String(),
			)
			return
		}
	}

	return
}
