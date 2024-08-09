package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GenesisRefundBid refunds the bid in genesis initialization.
// This action will mint coins to the module account and send coins to the bidder.
// The reason for minting is that the module account has no balance during genesis initialization.
func (k Keeper) GenesisRefundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid) error {
	soBid.Params = nil // treat it as refund name orders
	return k.refundBid(ctx, soBid, dymnstypes.NameOrder, true)
}

// RefundBid refunds the bid.
// This action will send coins from module account to the bidder.
func (k Keeper) RefundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid, orderType dymnstypes.OrderType) error {
	return k.refundBid(ctx, soBid, orderType, false)
}

// refundBid refunds the bid.
// Depends on the genesis flag, this action will mint coins to the module account and send coins to the bidder.
func (k Keeper) refundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid, orderType dymnstypes.OrderType, genesis bool) error {
	if err := soBid.Validate(orderType); err != nil {
		return err
	}

	if genesis {
		// During genesis initialization progress, the module account has no balance, so we mint coins.
		// Otherwise, the module account should have enough balance to refund the bid.
		if err := k.bankKeeper.MintCoins(ctx, dymnstypes.ModuleName, sdk.Coins{soBid.Price}); err != nil {
			return err
		}
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		dymnstypes.ModuleName,
		sdk.MustAccAddressFromBech32(soBid.Bidder),
		sdk.Coins{soBid.Price},
	); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			dymnstypes.EventTypeSoRefundBid,
			sdk.NewAttribute(dymnstypes.AttributeKeySoRefundBidder, soBid.Bidder),
			sdk.NewAttribute(dymnstypes.AttributeKeySoRefundAmount, soBid.Price.String()),
		),
	)

	return nil
}

// GenesisRefundBuyOrder refunds the buy orders in genesis initialization.
// This action will mint coins to the module account and send coins to the buyer.
// The reason for minting is that the module account has no balance during genesis initialization.
func (k Keeper) GenesisRefundBuyOrder(ctx sdk.Context, offer dymnstypes.BuyOrder) error {
	return k.refundBuyOrder(ctx, offer, true)
}

// RefundBuyOrder refunds the deposited amount to the buy order.
// This action will send coins from module account to the buyer.
func (k Keeper) RefundBuyOrder(ctx sdk.Context, offer dymnstypes.BuyOrder) error {
	return k.refundBuyOrder(ctx, offer, false)
}

// refundBuyOrder refunds the buy order.
// Depends on the genesis flag, this action will mint coins to the module account and send coins to the buyer.
func (k Keeper) refundBuyOrder(ctx sdk.Context, offer dymnstypes.BuyOrder, genesis bool) error {
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
			dymnstypes.EventTypeBoRefundOffer,
			sdk.NewAttribute(dymnstypes.AttributeKeyBoRefundBuyer, offer.Buyer),
			sdk.NewAttribute(dymnstypes.AttributeKeyBoRefundAmount, offer.OfferPrice.String()),
		),
	)

	return nil
}
