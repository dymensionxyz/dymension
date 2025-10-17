package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// BeginBlock iterates over auctions and checks for finished ones and interval pumping
func (k Keeper) BeginBlock(ctx sdk.Context) error {
	// Get all auctions from the store
	auctions, err := k.GetAllAuctions(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to get all auctions: %w", err)
	}

	// Iterate through auctions
	for _, auction := range auctions {
		if auction.Completed {
			continue
		}

		// Check if auction is completed (either due to time or being fully sold)
		// If completed => end auction (it creates a final pump stream)
		// If active => check if needs pumping
		if auction.EndTime.Before(ctx.BlockTime()) {
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
		} else if auction.IsActive(ctx.BlockTime()) {
			err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
				return k.ProcessIntervalPumping(ctx, auction)
			})
			if err != nil {
				k.Logger(ctx).Error("failed to process interval pumping",
					"auction_id", auction.Id,
					"error", err)
				continue
			}
		}
	}

	// Update the moving average prices for all accepted tokens
	k.UpdateAveragePrices(ctx)

	return nil
}

// UpdateAveragePrices updates the moving average prices for all accepted tokens
func (k Keeper) UpdateAveragePrices(ctx sdk.Context) {
	var denoms []string

	// Collect keys first
	err := k.acceptedTokens.Walk(ctx, nil, func(denom string, _ types.TokenData) (bool, error) {
		denoms = append(denoms, denom)
		return false, nil
	})
	if err != nil {
		k.Logger(ctx).Error("failed to collect denoms", "error", err)
		return
	}

	for _, denom := range denoms {
		err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.UpdateMovingAveragePrice(ctx, denom)
		})
		if err != nil {
			k.Logger(ctx).Error("failed to update moving average price", "denom", denom, "error", err)
		}
	}
}

func (k Keeper) UpdateMovingAveragePrice(ctx sdk.Context, denom string) error {
	tokenData, err := k.GetAcceptedTokenData(ctx, denom)
	if err != nil {
		return err
	}

	// get current price from amm
	price, err := k.ammKeeper.CalculateSpotPrice(ctx, tokenData.PoolId, denom, k.baseDenom)
	if err != nil {
		return err
	}

	// EMA formula: new_avg = alpha * current_price + (1 - alpha) * old_avg
	alpha := k.MustGetParams(ctx).MovingAverageSmoothingFactor
	oneMinusAlpha := math.LegacyOneDec().Sub(alpha)
	newAverage := alpha.Mul(price).Add(oneMinusAlpha.Mul(tokenData.LastAveragePrice))

	tokenData.LastAveragePrice = newAverage
	err = k.SetAcceptedToken(ctx, denom, tokenData)
	if err != nil {
		return err
	}

	return nil
}
