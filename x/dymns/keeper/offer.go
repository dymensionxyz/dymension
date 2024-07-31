package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GenesisRefundOffer refunds the offer-to-buy in genesis initialization.
// This action will mint coins to the module account and send coins to the buyer.
func (k Keeper) GenesisRefundOffer(ctx sdk.Context, offer dymnstypes.OfferToBuy) error {
	return k.refundOffer(ctx, offer, true)
}

// RefundOffer refunds the offer-to-buy.
// This action will send coins from module account to the buyer.
func (k Keeper) RefundOffer(ctx sdk.Context, offer dymnstypes.OfferToBuy) error {
	return k.refundOffer(ctx, offer, false)
}

func (k Keeper) refundOffer(ctx sdk.Context, offer dymnstypes.OfferToBuy, genesis bool) error {
	if err := offer.Validate(); err != nil {
		return err
	}

	if genesis {
		// During genesis initialization progress, the module account has no balance, so we mint coins.
		// Otherwise, the module account should have enough balance to refund the offer.
		if err := k.bankKeeper.MintCoins(ctx, dymnstypes.ModuleName, sdk.Coins{offer.OfferPrice}); err != nil {
			return err
		}
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(offer.Buyer),
		sdk.Coins{offer.OfferPrice},
	); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			dymnstypes.EventTypeOtbRefundOffer,
			sdk.NewAttribute(dymnstypes.AttributeKeyOtbRefundBuyer, offer.Buyer),
			sdk.NewAttribute(dymnstypes.AttributeKeyOtbRefundAmount, offer.OfferPrice.String()),
		),
	)

	return nil
}
