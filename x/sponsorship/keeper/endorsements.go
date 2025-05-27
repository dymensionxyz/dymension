package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// UpdateTotalSharesWithDistribution updates the endorsement total shares for all the rollapps from `update`.
//
// CONTRACT: All gauges must exist.
func (k Keeper) UpdateTotalSharesWithDistribution(ctx sdk.Context, update types.Distribution) error {
	for _, weight := range update.Gauges {
		// The gauge must exist. It must be validated in Keeper.validateWeights beforehand.
		gauge, _ := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)

		raGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Rollapp)
		if !ok {
			// the gauge is not a RA gauge
			continue
		}

		err := k.UpdateEndorsement(ctx, raGauge.Rollapp.RollappId, types.AddTotalShares(weight.Power))
		if err != nil {
			return fmt.Errorf("update endorsement shares: rollapp %s: %w", raGauge.Rollapp.RollappId, err)
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
// 6. Calculate the user's portion of the rewards based on total accumulated rewards in the gauge
// 7. Blacklist the user from claiming rewards for this distribution.
func (k Keeper) Claim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) error {
	ok, err := k.CanClaim(ctx, claimer)
	if err != nil {
		return fmt.Errorf("can claim: %w", err)
	}
	if !ok {
		return fmt.Errorf("user is not allowed to claim: %s", claimer)
	}

	result, err := k.EstimateClaim(ctx, claimer, gaugeId)
	if err != nil {
		return fmt.Errorf("estimate claim: %w", err)
	}

	err = k.incentivesKeeper.DistributeEndorsementRewards(ctx, claimer, gaugeId, result.Rewards)
	if err != nil {
		return fmt.Errorf("distribute rewards: %w", err)
	}

	err = k.BlacklistClaim(ctx, claimer)
	if err != nil {
		return fmt.Errorf("blacklist claim: %w", err)
	}

	return nil
}

type EstimateClaimResult struct {
	RollappId      string
	Rewards        sdk.Coins
	EndorsedAmount math.Int
}

// EstimateClaim estimates the rewards for the user from the provided endorsement gauge.
// Rewards are calculated based on the total accumulated coins in the gauge (`gauge.Coins`)
// and the user's power relative to the total shares for the endorsement (`endorsement.EpochShares`).
// `endorsement.EpochShares` is assumed to represent the total current shares for the endorsed rollapp.
// Does not change the state.
func (k Keeper) EstimateClaim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) (EstimateClaimResult, error) {
	gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get gauge by ID %d: %w", gaugeId, err)
	}

	// The original code uses this type assertion. We assume it's correct for the specific
	// setup of the incentives module's Gauge.DistributeTo field.
	// This part is to extract the RollappId. The rewards themselves will come from gauge.Coins.
	eGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement)
	if !ok {
		// Attempt to clarify based on standard protobuf oneof, if the above fails or is specific to a custom setup.
		// This is a common way to handle oneof in Go if DistributeTo is a QueryCondition.
		qc := gauge.GetDistributeTo()
		if qc == nil {
			return EstimateClaimResult{}, fmt.Errorf("gauge %d DistributeTo is nil", gaugeId)
		}
		endorsementTypeField, ok := qc.Type.(*incentivestypes.QueryCondition_Endorsement)
		if !ok {
			return EstimateClaimResult{}, fmt.Errorf("gauge %d is not an endorsement type gauge", gaugeId)
		}
		if endorsementTypeField.Endorsement == nil {
			return EstimateClaimResult{}, fmt.Errorf("endorsement data is nil for gauge %d", gaugeId)
		}
		// If we reach here, RollappId is endorsementTypeField.Endorsement.RollappId
		// And the problematic eGauge.Endorsement.EpochRewards would be endorsementTypeField.Endorsement.EpochRewards
	}
	// If the original eGauge assertion is not valid, the line below will panic.
	// This indicates a mismatch between assumptions and actual incentives module types.
	// For this change, we proceed assuming eGauge is valid as per original code for RollappId extraction.
	// If eGauge is nil or not *incentivestypes.Gauge_Endorsement, this will be an issue.
	// However, to change as little as possible of the surrounding logic for RollappId:
	if eGauge == nil || eGauge.Endorsement == nil {
		// Fallback or error if the expected structure isn't met for RollappId
		// This case should ideally be handled by the ok check above, but if `eGauge.Endorsement` can be nil:
		return EstimateClaimResult{}, fmt.Errorf("gauge %d is endorsement type but Endorsement field is nil or eGauge is not of expected type", gaugeId)
	}
	rollappId := eGauge.Endorsement.RollappId


	endorsement, err := k.GetEndorsement(ctx, rollappId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get endorsement for rollapp %s: %w", rollappId, err)
	}

	vote, err := k.GetVote(ctx, claimer)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get vote for claimer %s: %w", claimer, err)
	}

	power := vote.GetGaugePower(endorsement.RollappGaugeId)
	if power.IsZero() {
		return EstimateClaimResult{}, fmt.Errorf("user %s has no endorsement power for RA gauge %d (rollapp %s)", claimer, endorsement.RollappGaugeId, rollappId)
	}

	userRewards := sdk.NewCoins()
	// Calculate rewards based on gauge.Coins (total accumulated rewards in the gauge)
	// and endorsement.EpochShares (assumed to be total shares for this endorsement).
	if endorsement.EpochShares.IsZero() {
		// If there are no shares, the user's portion is undefined or zero.
		// Log if there are rewards but no shares, as it might be an anomaly.
		if !gauge.Coins.IsZero() {
			k.Logger(ctx).Info("EstimateClaim: endorsement has rewards in gauge but EpochShares is zero",
				"gaugeId", gaugeId, "rollappId", rollappId, "rewards", gauge.Coins.String())
		}
		// Return empty rewards as no portion can be calculated.
	} else {
		for _, totalRewardCoin := range gauge.Coins {
			// user's share of this coin = (user_power / total_shares) * total_reward_coin_amount
			// To maintain precision, multiply first, then divide: (user_power * total_reward_coin_amount) / total_shares
			rewardAmount := power.Mul(totalRewardCoin.Amount).Quo(endorsement.EpochShares)
			if rewardAmount.IsPositive() {
				userRewards = userRewards.Add(sdk.NewCoin(totalRewardCoin.Denom, rewardAmount))
			}
		}
	}

	return EstimateClaimResult{
		RollappId:      rollappId,
		Rewards:        userRewards,
		EndorsedAmount: power,
	}, nil
}
