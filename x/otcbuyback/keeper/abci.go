package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// BeginBlock iterates over auctions and checks for finished ones
func (k Keeper) BeginBlock(ctx sdk.Context) error {
	// Get all auctions from the store
	auctions, err := k.GetAllAuctions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all auctions: %w", err)
	}

	// Iterate through auctions and check for completed ones
	for _, auction := range auctions {
		// Check if auction is completed (either due to time or being fully sold)
		if !auction.Completed && auction.EndTime.Before(ctx.BlockTime()) {

			k.Logger(ctx).Info("processing completed auction",
				"auction_id", auction.Id,
				"end_time", auction.EndTime,
				"block_time", ctx.BlockTime(),
				"sold_amount", auction.SoldAmount,
				"allocation", auction.Allocation)

			// Process the completed auction
			err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return k.EndAuction(ctx, auction.Id, "auction_completed")
			})
			if err != nil {
				k.Logger(ctx).Error("failed to end auction",
					"auction_id", auction.Id,
					"error", err)
				continue
			}
		}
	}

	// Update the TWAPs for all accepted tokens
	err = k.UpdateTWAPs(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to update TWAPs", "error", err)
		// FIXME: maybe we should halt the the auction in this case??
	}

	return nil
}

// UpdateTWAPs updates the TWAPs for all accepted tokens
func (k Keeper) UpdateTWAPs(ctx sdk.Context) error {
	// FIXME: TODO
	return nil
}
