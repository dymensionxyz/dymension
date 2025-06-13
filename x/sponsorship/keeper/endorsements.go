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
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return fmt.Errorf("has endorser position: %w", err)
		}
		if errors.Is(err, collections.ErrNotFound) {
			// Must initialize endorser shares with zero to avoid panic
			endorserPosition = types.NewDefaultEndorserPosition()
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
// 1. Get the endorsement gauge by gaugeId
// 2. Get associated rollappId from the endorsement gauge
// 3. Get endorsement by rollappId
// 4. Get rollapp gauge associated with the endorsement
// 5. Get the user's power cast for the rollapp gauge
// 6. Calculate the user's portion of the rewards
// 7. Update the endorsement epoch shares
func (k Keeper) Claim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) error {
	result, err := k.EstimateClaim(ctx, claimer, gaugeId)
	if err != nil {
		return fmt.Errorf("estimate claim: %w", err)
	}

	if result.Rewards.IsZero() {
		// Nothing to claim
		return nil
	}

	// Rewards reside in x/incentives module
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, incentivestypes.ModuleName, claimer, result.Rewards)
	if err != nil {
		return fmt.Errorf("send coins from x/incentives to user: %w", err)
	}

	endorsement, err := k.GetEndorsement(ctx, result.RollappId)
	if err != nil {
		return fmt.Errorf("get endorsement: %w", err)
	}

	endorsement.DistributedCoins = endorsement.DistributedCoins.Add(result.Rewards...)

	err = k.SaveEndorsement(ctx, endorsement)
	if err != nil {
		return fmt.Errorf("save endorsement: %w", err)
	}

	endorserPosition, err := k.GetEndorserPosition(ctx, claimer, result.RollappId)
	if err != nil {
		return fmt.Errorf("get endorser position: %w", err)
	}

	endorserPosition.LastSeenAccumulator = endorsement.Accumulator
	endorserPosition.AccumulatedRewards = sdk.NewCoins()

	err = k.SaveEndorserPosition(ctx, claimer, result.RollappId, endorserPosition)
	if err != nil {
		return fmt.Errorf("save endorser position: %w", err)
	}

	return nil
}

type EstimateClaimResult struct {
	RollappId string
	Rewards   sdk.Coins
}

// EstimateClaim estimates the rewards for the user from the provided endorsement gauge.
// Does not change the state.
func (k Keeper) EstimateClaim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) (EstimateClaimResult, error) {
	gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get gauge: %w", err)
	}

	raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Rollapp)
	if !ok {
		return EstimateClaimResult{}, fmt.Errorf("gauge is not rollapp: %d", gaugeId)
	}

	endorsement, err := k.GetEndorsement(ctx, raGauge.Rollapp.RollappId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get endorsement: %w", err)
	}

	endorserPosition, err := k.GetEndorserPosition(ctx, claimer, raGauge.Rollapp.RollappId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get endorser position: %w", err)
	}

	// Calculate newly accrued rewards
	// RewardsToBank truncates the decimal part of the rewards. They will accumulate as dust in x/incentives.
	newlyAccruedRewardsDec := endorserPosition.RewardsToBank(endorsement.Accumulator)

	// Total rewards to claim are newly accrued rewards plus any previously accumulated rewards
	totalRewardsToClaim := newlyAccruedRewardsDec.Add(endorserPosition.AccumulatedRewards...)

	return EstimateClaimResult{
		RollappId: raGauge.Rollapp.RollappId,
		Rewards:   totalRewardsToClaim,
	}, nil
}

// UpdateEndorsementTotalCoins updates the total coins for an endorsement by adding the provided coins.
// This is used when new rewards are added to the endorsement gauge.
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
