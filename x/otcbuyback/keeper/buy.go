package keeper

import (
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// Buy allows a user to purchase specific amount of tokens in an active auction
func (k Keeper) Buy(
	ctx sdk.Context,
	buyer sdk.AccAddress,
	auctionID uint64,
	amountToBuy math.Int,
	denomToPay string,
) (sdk.Coin, error) {

	// Get and validate auction
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.Coin{}, types.ErrAuctionNotFound
	}

	// Check if auction is active
	if !auction.IsActive(ctx.BlockTime()) {
		return sdk.Coin{}, types.ErrAuctionNotActive
	}

	// Validate payment token is accepted
	allowedDenoms := k.MustGetParams(ctx).AcceptedTokens
	if !slices.Contains(allowedDenoms, denomToPay) {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrTokenNotAccepted,
			"token %s not accepted for this auction", denomToPay)
	}

	// Check if enough tokens are available
	remainingAllocation := auction.GetRemainingAllocation()
	if amountToBuy.GT(remainingAllocation) {
		return sdk.Coin{}, types.ErrInsufficientAllocation
	}

	// Get current price
	currentPrice, err := k.GetCurrentPrice(ctx, auctionID, denomToPay, ctx.BlockTime())
	if err != nil {
		return sdk.Coin{}, err
	}

	// Calculate tokens that can be purchased
	paymentAmt := math.LegacyNewDecFromInt(amountToBuy).Mul(currentPrice).TruncateInt()
	if paymentAmt.IsZero() {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrInvalidPurchaseAmount,
			"payment amount too small to purchase any tokens")
	}
	paymentCoin := sdk.NewCoin(denomToPay, paymentAmt)

	// Transfer payment from buyer to auction module account
	err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		buyer,
		types.ModuleName,
		sdk.NewCoins(paymentCoin),
	)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to transfer payment")
	}

	// Update or create purchase record
	purchase, found := k.GetPurchase(ctx, auctionID, buyer.String())
	if found {
		// Update existing purchase
		purchase.Amount = purchase.Amount.Add(amountToBuy)
	} else {
		// Create new purchase record
		purchase = types.NewUserVestingPlan(
			amountToBuy,
			auction.GetVestingStartTime(),
			auction.GetVestingEndTime(),
		)
	}

	// Save purchase
	if err := k.SetPurchase(ctx, auctionID, buyer.String(), purchase); err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to save purchase")
	}

	// Update auction totals
	auction.SoldAmount = auction.SoldAmount.Add(amountToBuy)
	auction.RaisedAmount = auction.RaisedAmount.Add(paymentCoin)

	// Save updated auction
	k.SetAuction(ctx, auction)

	// Emit purchase event
	err = uevent.EmitTypedEvent(ctx, &types.EventTokensPurchased{
		AuctionId:       auctionID,
		Buyer:           buyer.String(),
		TokensPurchased: amountToBuy,
		AmountPaid:      paymentCoin,
		PricePerToken:   currentPrice,
	})
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to emit purchase event")
	}

	k.Logger().Info("tokens purchased",
		"auction_id", auctionID,
		"buyer", buyer,
		"tokens", amountToBuy,
		"payment", paymentCoin,
		"price", currentPrice)

	// Check if auction should end (sold out)
	if auction.SoldAmount.GTE(auction.Allocation.Amount) {
		auction.EndTime = ctx.BlockTime()

		// FIXME: call end auction
	}

	return paymentCoin, nil

}

// BuyExactSpend allows a user to purchase tokens in an active auction
func (k Keeper) BuyExactSpend(
	ctx sdk.Context,
	buyer sdk.AccAddress,
	auctionID uint64,
	paymentCoin sdk.Coin,
) (math.Int, error) {
	// Get current price
	currentPrice, err := k.GetCurrentPrice(ctx, auctionID, paymentCoin.Denom, ctx.BlockTime())
	if err != nil {
		return math.ZeroInt(), err
	}

	// Calculate tokens that can be purchased
	tokensToPurchase := math.LegacyNewDecFromInt(paymentCoin.Amount).Quo(currentPrice).TruncateInt()
	if tokensToPurchase.IsZero() {
		return math.ZeroInt(), errorsmod.Wrap(types.ErrInvalidPurchaseAmount,
			"payment amount too small to purchase any tokens")
	}

	_, err = k.Buy(ctx, buyer, auctionID, tokensToPurchase, paymentCoin.Denom)
	if err != nil {
		return math.ZeroInt(), err
	}

	return tokensToPurchase, nil
}

// GetCurrentPrice returns the current price for an active auction
func (k Keeper) GetCurrentPrice(ctx sdk.Context, auctionID uint64, quoteDenom string, currentTime time.Time) (math.LegacyDec, error) {

	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return math.LegacyZeroDec(), types.ErrAuctionNotFound
	}
	baseDenom := auction.Allocation.Denom

	// FIXME: Get pool ID for quote denom
	var poolID uint64

	// Get base price
	// FIXME: wrap with TWAP logic
	base_price, err := k.ammKeeper.CalculateSpotPrice(ctx, poolID, quoteDenom, baseDenom)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	discount := auction.GetCurrentDiscount(currentTime)

	// Price = AMM Price × (1 - Current Discount%)
	discountMultiplier := math.LegacyOneDec().Sub(discount)
	return base_price.Mul(discountMultiplier), nil
}
