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
	pumpParams types.Auction_PumpParams,
) (uint64, error) {
	if allocation.Denom != k.baseDenom {
		return 0, errorsmod.Wrap(gerrc.ErrInvalidArgument, "allocation must be in base denom")
	}

	expectedBalance := math.ZeroInt()
	err := k.auctions.Walk(ctx, nil, func(key uint64, auction types.Auction) (bool, error) {
		if auction.IsCompleted() {
			return false, nil
		}
		expectedBalance = expectedBalance.Add(auction.GetRemainingAllocation())
		return false, nil
	})
	if err != nil {
		return 0, err
	}

	// add the new allocation to the already allocated
	expectedBalance = expectedBalance.Add(allocation.Amount)

	// check if the module account has enough funds
	bankBalance := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), k.baseDenom)
	if bankBalance.Amount.LT(expectedBalance) {
		return 0, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "insufficient funds")
	}

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
		pumpParams,
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
	auction.EndTime = ctx.BlockTime()
	auction.Completed = true

	err := k.SetAuction(ctx, auction)
	if err != nil {
		return errorsmod.Wrap(err, "failed to set auction")
	}

	// create pump streams
	pumpStreams, err := k.CreatePumpStreams(ctx, auction)
	if err != nil {
		return errorsmod.Wrap(err, "failed to create pump streams")
	}

	// Emit completion event
	err = uevent.EmitTypedEvent(ctx, &types.EventAuctionCompleted{
		AuctionId:        auctionID,
		TotalSold:        auction.SoldAmount,
		TotalRaised:      auction.RaisedAmount,
		PumpStreams:      pumpStreams,
		CompletionReason: reason,
	})
	if err != nil {
		return errorsmod.Wrap(err, "failed to emit auction completed event")
	}

	return nil
}

// CreateStream creates a pump stream struct given the required params.
func (k Keeper) CreatePumpStreams(ctx sdk.Context, auction types.Auction) ([]uint64, error) {
	var streamIDs []uint64

	coins := auction.RaisedAmount
	pp := auction.PumpParams

	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		streamertypes.ModuleName,
		coins,
	)
	if err != nil {
		return nil, err
	}

	// for each coin
	for _, coin := range coins {
		poolID, err := k.GetAcceptedTokenPoolID(ctx, coin.Denom)
		if err != nil {
			return nil, err
		}

		streamID, err := k.streamerKeeper.CreatePumpStream(ctx,
			streamertypes.CreateStreamGeneric{
				Coins:             sdk.NewCoins(coin),
				StartTime:         ctx.BlockTime().Add(pp.StartTimeAfterAuctionEnd),
				EpochIdentifier:   pp.EpochIdentifier,
				NumEpochsPaidOver: pp.NumEpochs,
			},
			pp.NumOfPumpsPerEpoch,
			pp.PumpDistr,
			true,
			streamertypes.PumpTargetPool(poolID, k.baseDenom),
		)
		if err != nil {
			return nil, err
		}
		streamIDs = append(streamIDs, streamID)
	}

	return streamIDs, nil
}
