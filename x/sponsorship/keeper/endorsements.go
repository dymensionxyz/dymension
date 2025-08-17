package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// UpdateEndorsementsAndPositions updates the endorsement and endorser positions based
// on the provided weight update.
//
// CONTRACT: All gauges must exist.
func (k Keeper) UpdateEndorsementsAndPositions(
	ctx sdk.Context,
	voter sdk.AccAddress,
	weights types.Distribution,
) error {
	for _, weight := range weights.Gauges {
		// The gauge must exist. It must be validated in Keeper.validateWeights beforehand.
		gauge, _ := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)

		raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Rollapp)
		if !ok {
			// the gauge is not a RA gauge
			continue
		}
		// If gauge is RA gauge, then we can extract associated rollapp ID
		var (
			raID = raGauge.Rollapp.RollappId
			// endorser's voting power cast to this rollapp
			shares = math.LegacyNewDecFromInt(weight.Power)
		)

		endorsement, err := k.GetEndorsement(ctx, raID)
		if err != nil {
			return fmt.Errorf("get endorsement: %w", err)
		}

		// Update total shares for this rollapp
		endorsement.TotalShares = endorsement.TotalShares.Add(shares)

		endorserPosition, err := k.GetEndorserPosition(ctx, voter, raID)
		if errors.Is(err, collections.ErrNotFound) {
			// Must initialize endorser shares with zero to avoid panic
			endorserPosition = types.NewDefaultEndorserPosition()
		} else if err != nil {
			return fmt.Errorf("get endorser position: %w", err)
		}

		// RewardsToBank truncates the decimal part of the rewards. They will accumulate as dust in x/incentives.
		rewardsToBank := endorserPosition.RewardsToBank(endorsement.Accumulator)

		// Update endorser position
		endorserPosition.Shares = endorserPosition.Shares.Add(shares)
		endorserPosition.LastSeenAccumulator = endorsement.Accumulator
		endorserPosition.AccumulatedRewards = endorserPosition.AccumulatedRewards.Add(rewardsToBank...)

		err = k.SaveEndorsement(ctx, endorsement)
		if err != nil {
			return fmt.Errorf("save endorsement: %w", err)
		}

		err = k.SaveEndorserPosition(ctx, voter, raID, endorserPosition)
		if err != nil {
			return fmt.Errorf("save endorser position: %w", err)
		}
	}
	return nil
}

// Claim claims the rewards for the user from the provided endorsement gauge.
// 1. Get endorsement by rollappId
// 2. Get user endorsement position
// 3. Estimate user rewards based on accumulators
// 4. Update endorsement and user position
func (k Keeper) Claim(ctx sdk.Context, claimer sdk.AccAddress, rollappId string) error {
	rewards, err := k.EstimateClaim(ctx, claimer, rollappId)
	if err != nil {
		return fmt.Errorf("estimate claim: %w", err)
	}

	if rewards.IsZero() {
		// Nothing to claim
		return nil
	}

	// Rewards reside in x/incentives module
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, incentivestypes.ModuleName, claimer, rewards)
	if err != nil {
		return fmt.Errorf("send coins from x/incentives to user: %w", err)
	}

	endorsement, err := k.GetEndorsement(ctx, rollappId)
	if err != nil {
		return fmt.Errorf("get endorsement: %w", err)
	}

	endorsement.DistributedCoins = endorsement.DistributedCoins.Add(rewards...)

	err = k.SaveEndorsement(ctx, endorsement)
	if err != nil {
		return fmt.Errorf("save endorsement: %w", err)
	}

	endorserPosition, err := k.GetEndorserPosition(ctx, claimer, rollappId)
	if err != nil {
		return fmt.Errorf("get endorser position: %w", err)
	}

	endorserPosition.LastSeenAccumulator = endorsement.Accumulator
	endorserPosition.AccumulatedRewards = sdk.NewCoins()

	err = k.SaveEndorserPosition(ctx, claimer, rollappId, endorserPosition)
	if err != nil {
		return fmt.Errorf("save endorser position: %w", err)
	}

	return nil
}

// EstimateClaim estimates user rewards for the given rollapp.
// Does not change the state.
func (k Keeper) EstimateClaim(ctx sdk.Context, claimer sdk.AccAddress, rollappId string) (sdk.Coins, error) {
	endorsement, err := k.GetEndorsement(ctx, rollappId)
	if err != nil {
		return nil, fmt.Errorf("get endorsement: %w", err)
	}

	endorserPosition, err := k.GetEndorserPosition(ctx, claimer, rollappId)
	if err != nil {
		return nil, fmt.Errorf("get endorser position: %w", err)
	}

	// Calculate newly accrued rewards
	// RewardsToBank truncates the decimal part of the rewards. They will accumulate as dust in x/incentives.
	newlyAccruedRewardsDec := endorserPosition.RewardsToBank(endorsement.Accumulator)

	// Total rewards to claim are newly accrued rewards plus any previously accumulated rewards
	totalRewardsToClaim := newlyAccruedRewardsDec.Add(endorserPosition.AccumulatedRewards...)

	return totalRewardsToClaim, nil
}

// UpdateEndorsementTotalCoins updates the total coins for an endorsement by adding the provided coins.
// This is used when new rewards are allocated by the endorsement gauge.
// It also updates the lazy accumulator by calculating the rewards per share.
func (k Keeper) UpdateEndorsementTotalCoins(ctx sdk.Context, rollappID string, additionalCoins sdk.Coins) error {
	endorsement, err := k.GetEndorsement(ctx, rollappID)
	if err != nil {
		return fmt.Errorf("get endorsement: %w", err)
	}

	// Update the lazy accumulator: add rewards per share to the accumulator
	// Only update if there are shares to avoid division by zero
	if endorsement.TotalShares.IsZero() {
		return types.ErrNoEndorsers
	}

	additionalDecCoins := sdk.NewDecCoinsFromCoins(additionalCoins...)
	// It is important to use QuoDecTruncate instead of Quo to avoid rounding errors
	// This ensures that claimable rewards are always less than or equal to rewards added to the gauge
	rewardsPerShare := additionalDecCoins.QuoDecTruncate(endorsement.TotalShares)
	endorsement.Accumulator = endorsement.Accumulator.Add(rewardsPerShare...)
	endorsement.TotalCoins = endorsement.TotalCoins.Add(additionalCoins...)

	err = k.SaveEndorsement(ctx, endorsement)
	if err != nil {
		return fmt.Errorf("save endorsement: %w", err)
	}

	return nil
}
