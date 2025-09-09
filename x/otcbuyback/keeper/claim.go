package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// ClaimVestedTokens allows users to claim their vested tokens
func (k Keeper) ClaimVestedTokens(ctx sdk.Context, claimer sdk.AccAddress, auctionID uint64) (math.Int, error) {
	// Get auction
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return math.ZeroInt(), types.ErrAuctionNotFound
	}

	// Only allow claiming from completed auctions
	if !auction.IsCompleted(ctx.BlockTime()) {
		return math.ZeroInt(), errorsmod.Wrap(gerrc.ErrFailedPrecondition, "auction must be completed to claim tokens")
	}

	// Get user's purchase
	purchase, found := k.GetPurchase(ctx, auctionID, claimer.String())
	if !found {
		return math.ZeroInt(), types.ErrNoUserPurchaseFound
	}

	// Calculate claimable amount
	claimableAmount := purchase.VestedAmount(ctx.BlockTime())
	if claimableAmount.IsZero() {
		return math.ZeroInt(), types.ErrNoClaimableTokens
	}

	// Update vesting plan
	purchase.ClaimTokens(claimableAmount)

	// Save updated purchase
	if err := k.SetPurchase(ctx, auctionID, claimer.String(), purchase); err != nil {
		return math.ZeroInt(), errorsmod.Wrap(err, "failed to save purchase")
	}

	// Transfer tokens from auction module to claimer
	claimCoin := sdk.NewCoin(auction.Allocation.Denom, claimableAmount)
	err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		claimer,
		sdk.NewCoins(claimCoin),
	)
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrap(err, "failed to transfer tokens")
	}

	// Emit claim event
	err = uevent.EmitTypedEvent(ctx, &types.EventTokensClaimed{
		AuctionId:        auctionID,
		Claimer:          claimer.String(),
		ClaimedAmount:    claimableAmount,
		RemainingVesting: purchase.GetRemainingVesting(),
	})
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrap(err, "failed to emit claim event")
	}

	k.Logger().Info("tokens claimed",
		"auction_id", auctionID,
		"claimer", claimer,
		"amount", claimableAmount)

	return claimableAmount, nil
}
