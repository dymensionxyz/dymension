package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
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
	vestingPeriod time.Duration,
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
	if !k.IsAcceptedDenom(ctx, denomToPay) {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrTokenNotAccepted,
			"token %s not accepted for this auction", denomToPay)
	}

	// Get params for validation
	params, err := k.GetParams(ctx)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to get params")
	}

	// Check minimum purchase amount
	if amountToBuy.LT(params.MinPurchaseAmount) {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidPurchaseAmount,
			"purchase amount %s is less than minimum %s", amountToBuy, params.MinPurchaseAmount)
	}

	// Check if enough tokens are available
	remainingAllocation := auction.GetRemainingAllocation()
	if amountToBuy.GT(remainingAllocation) {
		return sdk.Coin{}, types.ErrInsufficientAllocation
	}

	// Get the current price and associated vesting period
	// For linear discount, it's hardcoded in the auction
	// For fixed discount, it's the same as specified in the request (if valid; otherwise, error)
	currentPrice, actualVestingPeriod, err := k.GetDiscountedPrice(ctx, auctionID, denomToPay, ctx.BlockTime(), vestingPeriod)
	if err != nil {
		return sdk.Coin{}, err
	}

	// Calculate tokens that can be purchased
	paymentAmt := math.LegacyNewDecFromInt(amountToBuy).Mul(currentPrice).TruncateInt()
	if !paymentAmt.IsPositive() {
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

	// Get or create purchase record
	purchase, found := k.GetPurchase(ctx, auctionID, buyer)
	if !found {
		purchase = types.NewPurchase()
	}

	// Enforce purchase limits (DoS protection)
	if uint64(len(purchase.Entries)) >= params.MaxPurchaseNumber {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrInvalidPurchaseAmount,
			fmt.Sprintf("maximum purchases per user per auction is %d", params.MaxPurchaseNumber))
	}

	// Create a new purchase entry with vesting start after delay
	purchase.AddEntry(types.NewPurchaseEntry(
		amountToBuy,
		auction.GetVestingStartTime(ctx.BlockTime()),
		actualVestingPeriod,
	))

	// Save purchase
	if err := k.SetPurchase(ctx, auctionID, buyer, purchase); err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to save purchase")
	}

	// Update auction totals
	auction.SoldAmount = auction.SoldAmount.Add(amountToBuy)
	auction.RaisedAmount = auction.RaisedAmount.Add(paymentCoin)

	// Save updated auction
	err = k.SetAuction(ctx, auction)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "failed to save updated auction")
	}

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

	k.Logger(ctx).Info("tokens purchased",
		"auction_id", auctionID,
		"buyer", buyer,
		"tokens", amountToBuy,
		"payment", paymentCoin,
		"price", currentPrice,
		"vesting_period", vestingPeriod)

	// Check if auction should end (sold out):
	// 'remaining = allocation - sold < min' => no one could buy the remaining part since it's too small.
	// If min is 0, the previous check is not performed, so we check 'allocation <= sold' instead.
	if auction.SoldAmount.GTE(auction.Allocation) || auction.GetRemainingAllocation().LT(params.MinPurchaseAmount) {
		// we make the end auction flow gas free, as it's not relevant to the specific user's action
		noGasCtx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		err = k.EndAuction(noGasCtx, auctionID, "auction_sold_out")
		if err != nil {
			return sdk.Coin{}, errorsmod.Wrap(err, "failed to end auction")
		}
	}

	return paymentCoin, nil
}

// BuyExactSpend allows a user to purchase tokens in an active auction
func (k Keeper) BuyExactSpend(
	ctx sdk.Context,
	buyer sdk.AccAddress,
	auctionID uint64,
	paymentCoin sdk.Coin,
	vestingPeriod time.Duration,
) (math.Int, error) {
	// Get current price
	currentPrice, _, err := k.GetDiscountedPrice(ctx, auctionID, paymentCoin.Denom, ctx.BlockTime(), vestingPeriod)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Calculate tokens that can be purchased
	tokensToPurchase := math.LegacyNewDecFromInt(paymentCoin.Amount).Quo(currentPrice).TruncateInt()
	if tokensToPurchase.IsZero() {
		return math.ZeroInt(), errorsmod.Wrap(types.ErrInvalidPurchaseAmount,
			"payment amount too small to purchase any tokens")
	}

	_, err = k.Buy(ctx, buyer, auctionID, tokensToPurchase, paymentCoin.Denom, vestingPeriod)
	if err != nil {
		return math.ZeroInt(), err
	}

	return tokensToPurchase, nil
}

// GetDiscountedPrice returns the current price for an active auction and associated vesting period
func (k Keeper) GetDiscountedPrice(
	ctx sdk.Context,
	auctionID uint64,
	quoteDenom string,
	currentTime time.Time,
	vestingPeriod time.Duration,
) (math.LegacyDec, time.Duration, error) {
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return math.LegacyZeroDec(), 0, types.ErrAuctionNotFound
	}

	// Get discount based on the auction type
	discount, actualVestingPeriod, err := auction.GetDiscount(currentTime, vestingPeriod)
	if err != nil {
		return math.LegacyZeroDec(), 0, fmt.Errorf("get discount: %w", err)
	}

	basePrice, err := k.GetBasePrice(ctx, quoteDenom)
	if err != nil {
		return math.LegacyZeroDec(), 0, fmt.Errorf("get base price: %w", err)
	}

	discountMultiplier := math.LegacyOneDec().Sub(discount)

	// Price = AMM Price × (1 - Current Discount%)
	return basePrice.Mul(discountMultiplier), actualVestingPeriod, nil
}

// GetBasePrice gets base price (max(current_price, average_price))
// we take the max, to avoid double discount in case the price is peaking
func (k Keeper) GetBasePrice(ctx sdk.Context, quoteDenom string) (math.LegacyDec, error) {
	poolID, err := k.GetAcceptedTokenPoolID(ctx, quoteDenom)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("get accepted token pool id: %w", err)
	}
	currPrice, err := k.ammKeeper.CalculateSpotPrice(ctx, poolID, quoteDenom, k.baseDenom)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("calculate spot price: %w", err)
	}
	avgPrice, err := k.GetMovingAveragePrice(ctx, quoteDenom)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("get moving average price: %w", err)
	}
	return math.LegacyMaxDec(currPrice, avgPrice), nil
}
