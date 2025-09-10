package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"

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
	vestingParams types.Auction_VestingParams,
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
		vestingParams,
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

	k.Logger(ctx).Info("auction created", "auction_id", auctionID, "allocation", allocation)

	return auctionID, nil
}

// EndAuction marks an auction as completed and processes the proceeds
func (k Keeper) EndAuction(ctx sdk.Context, auctionID uint64, reason string) error {
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return types.ErrAuctionNotFound
	}

	// Check if auction is not already completed
	if auction.IsCompleted() {
		return types.ErrAuctionCompleted
	}

	// Process any remaining allocation - return to treasury
	remainingAllocation := auction.GetRemainingAllocation()
	if remainingAllocation.IsPositive() {
		err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			streamertypes.ModuleName,
			sdk.NewCoins(sdk.NewCoin(k.baseDenom, remainingAllocation)),
		)
		if err != nil {
			return errorsmod.Wrap(err, "failed to send remaining allocation to treasury")
		}
	}

	// Set the auction as completed
	auction.Completed = true

	err := k.SetAuction(ctx, auction)
	if err != nil {
		return errorsmod.Wrap(err, "failed to set auction")
	}

	// If ended prematurely, we need to update the vesting start and end times
	if ctx.BlockTime().Before(auction.GetEndTime()) {
		// FIXME: IMPLEMENT. can we iterate over the collection and update inplace?
	}

	// FIXME: create pump streams

	// Emit completion event
	err = uevent.EmitTypedEvent(ctx, &types.EventAuctionCompleted{
		AuctionId:        auctionID,
		TotalSold:        auction.SoldAmount,
		TotalRaised:      auction.RaisedAmount,
		CompletionReason: reason,
	})
	if err != nil {
		return errorsmod.Wrap(err, "failed to emit auction completed event")
	}

	return nil
}
