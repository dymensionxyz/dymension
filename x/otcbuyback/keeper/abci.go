package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
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
	k.UpdateTWAPs(ctx)

	return nil
}

// UpdateTWAPs updates the TWAPs for all accepted tokens
func (k Keeper) UpdateTWAPs(ctx sdk.Context) {
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
		//ApplyFuncIfNoError
		err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.UpdateTWAP(ctx, denom)
		})
		if err != nil {
			k.Logger(ctx).Error("failed to update TWAPs", "error", err)
			// FIXME: maybe we should halt the the auction in this case??
		}
	}
}

func (k Keeper) UpdateTWAP(ctx sdk.Context, denom string) error {
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
