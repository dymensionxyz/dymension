package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// CreateAuction creates a new Dutch auction
func (k Keeper) CreateAuction(
	ctx sdk.Context,
	allocation sdk.Coin,
	startTime time.Time,
	endTime time.Time,
	initialDiscount math.LegacyDec,
	maxDiscount math.LegacyDec,
	vestingPlan types.Auction_VestingPlan,
) (uint64, error) {

	if allocation.Denom != k.baseDenom {
		return 0, errorsmod.Wrap(gerrc.ErrInvalidArgument, "allocation must be in base denom")
	}

	// FIXME: validate the funds available in the module account
	// FIXME: need to mark them as allocated

	// Get next auction ID
	auctionID, err := k.IncrementNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}

	// Create auction
	auction := types.NewAuction(
		auctionID,
		allocation.Amount,
		startTime,
		endTime,
		initialDiscount,
		maxDiscount,
		vestingPlan,
	)

	// Validate auction
	if err := auction.ValidateBasic(); err != nil {
		return 0, errorsmod.Wrap(err, "invalid auction parameters")
	}

	// Store auction
	err = k.SetAuction(ctx, auction)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to set auction")
	}

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventAuctionCreated{
		AuctionId:       auctionID,
		Allocation:      allocation,
		StartTime:       startTime.String(),
		EndTime:         endTime.String(),
		InitialDiscount: initialDiscount.String(),
		MaxDiscount:     maxDiscount.String(),
	})
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to emit auction created event")
	}

	k.Logger().Info("auction created", "auction_id", auctionID, "allocation", allocation)

	return auctionID, nil
}

// EndAuction marks an auction as completed and processes the proceeds
func (k Keeper) EndAuction(ctx sdk.Context, auctionID uint64, reason string) error {

	// FIXME: do we want to update the vesting start and end times?

	// FIXME: what to do with remaining unsold funds?

	/* FIXME: Implement */
	/*
		auction, found := k.GetAuction(ctx, auctionID)
		if !found {
			return types.ErrAuctionNotFound
		}

		if auction.IsCompleted(ctx.BlockTime()) || auction.IsCancelled() {
			return types.ErrAuctionCompleted
		}

		// Update auction status
		auction.Status = types.AUCTION_STATUS_COMPLETED
		k.SetAuction(ctx, auction)

		// Process proceeds - convert raised funds to DYM and send to treasury
		err := k.processAuctionProceeds(ctx, auction)
		if err != nil {
			return errorsmod.Wrap(types.ErrTreasuryOperation, err.Error())
		}

		// Set up vesting for all purchases
		err = k.setupVestingForAuction(ctx, auction)
		if err != nil {
			return errorsmod.Wrap(err, "failed to setup vesting")
		}

		// Calculate final price
		var finalPrice math.LegacyDec
		if auction.SoldAmount.IsPositive() {
			// Get the average price from all purchases
			finalPrice = k.calculateAveragePurchasePrice(ctx, auctionID)
		}

		// Emit completion event
		err = uevent.EmitTypedEvent(ctx, &types.EventAuctionCompleted{
			AuctionId:        auctionID,
			TotalSold:        auction.SoldAmount,
			TotalRaised:      auction.RaisedAmount,
			FinalPrice:       finalPrice,
			CompletionReason: reason,
		})
		if err != nil {
			return errorsmod.Wrap(err, "failed to emit auction completed event")
		}

		k.Logger().Info("auction ended", "auction_id", auctionID, "reason", reason, "total_sold", auction.SoldAmount)

	*/
	return nil
}
