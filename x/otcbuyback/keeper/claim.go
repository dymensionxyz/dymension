package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// ClaimVestedTokens allows users to claim their vested tokens
// In Phase 2: Users can claim at any time (not just after auction completion)
func (k Keeper) ClaimVestedTokens(ctx sdk.Context, claimer sdk.AccAddress, auctionID uint64) (math.Int, error) {
	// Get user's purchase
	purchase, found := k.GetPurchase(ctx, auctionID, claimer)
	if !found {
		return math.ZeroInt(), types.ErrNoUserPurchaseFound
	}

	// Calculate claimable amount
	claimableAmount := purchase.ClaimableAmount(ctx.BlockTime())
	if claimableAmount.IsZero() {
		return math.ZeroInt(), types.ErrNoClaimableTokens
	}

	// Update total claimed amount
	purchase.ClaimTokens(claimableAmount)

	// Save updated purchase
	if err := k.SetPurchase(ctx, auctionID, claimer, purchase); err != nil {
		return math.ZeroInt(), errorsmod.Wrap(err, "failed to save purchase")
	}

	// Transfer tokens from auction module to claimer
	claimCoin := sdk.NewCoin(k.baseDenom, claimableAmount)
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
		RemainingVesting: purchase.UnclaimedAmount(),
	})
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrap(err, "failed to emit claim event")
	}

	k.Logger(ctx).Info("tokens claimed",
		"auction_id", auctionID,
		"claimer", claimer,
		"amount", claimableAmount,
		"remaining", purchase.UnclaimedAmount())

	return claimableAmount, nil
}
