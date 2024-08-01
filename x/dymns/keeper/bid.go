package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GenesisRefundBid refunds the bid in genesis initialization.
// This action will mint coins to the module account and send coins to the bidder.
// The reason for minting is that the module account has no balance during genesis initialization.
func (k Keeper) GenesisRefundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid) error {
	return k.refundBid(ctx, soBid, true)
}

// RefundBid refunds the bid.
// This action will send coins from module account to the bidder.
func (k Keeper) RefundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid) error {
	return k.refundBid(ctx, soBid, false)
}

// refundBid refunds the bid.
// Depends on the genesis flag, this action will mint coins to the module account and send coins to the bidder.
func (k Keeper) refundBid(ctx sdk.Context, soBid dymnstypes.SellOrderBid, genesis bool) error {
	if err := soBid.Validate(); err != nil {
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
			dymnstypes.EventTypeDymNameRefundBid,
			sdk.NewAttribute(dymnstypes.AttributeKeyDymNameRefundBidder, soBid.Bidder),
			sdk.NewAttribute(dymnstypes.AttributeKeyDymNameRefundAmount, soBid.Price.String()),
		),
	)

	return nil
}
