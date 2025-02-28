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
// 6. Calculate the user's portion of the rewards
// 7. Update the endorsement epoch shares
// 8. Blacklist the user from claiming rewards in this epoch
func (k Keeper) Claim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) error {
	result, err := k.EstimateClaim(ctx, claimer, gaugeId)
	if err != nil {
		return fmt.Errorf("estimate claim: %w", err)
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, incentivestypes.ModuleName, claimer, result.Rewards)
	if err != nil {
		return fmt.Errorf("send coins from x/incentives to user: %w", err)
	}

	err = k.UpdateEndorsement(ctx, result.RollappId, types.AddEpochShares(result.EndorsedAmount.Neg()))
	if err != nil {
		return fmt.Errorf("update endorsement epoch shares: %w", err)
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
// Does not change the state.
func (k Keeper) EstimateClaim(ctx sdk.Context, claimer sdk.AccAddress, gaugeId uint64) (EstimateClaimResult, error) {
	ok, err := k.CanClaim(ctx, claimer)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("can claim: %w", err)
	}
	if !ok {
		return EstimateClaimResult{}, fmt.Errorf("user is not allowed to claim: %s", claimer)
	}

	gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, gaugeId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get gauge: %w", err)
	}

	eGauge, ok := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement)
	if !ok {
		return EstimateClaimResult{}, fmt.Errorf("gauge is not endorsement: %d", gaugeId)
	}

	endorsement, err := k.GetEndorsement(ctx, eGauge.Endorsement.RollappId)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get endorsement: %w", err)
	}

	vote, err := k.GetVote(ctx, claimer)
	if err != nil {
		return EstimateClaimResult{}, fmt.Errorf("get vote: %w", err)
	}

	power := vote.GetGaugePower(endorsement.RollappGaugeId)
	if power.IsZero() {
		return EstimateClaimResult{}, fmt.Errorf("user does not endorse respective RA gauge: %d", gaugeId)
	}

	// Which portion of the rewards the user is entitled to
	userPortion := power.Quo(endorsement.EpochShares)

	var userRewards sdk.Coins
	for _, reward := range eGauge.Endorsement.EpochRewards {
		userRewards = append(userRewards, sdk.Coin{
			Denom:  reward.Denom,
			Amount: userPortion.Mul(reward.Amount),
		})
	}

	return EstimateClaimResult{
		RollappId:      eGauge.Endorsement.RollappId,
		Rewards:        userRewards,
		EndorsedAmount: power,
	}, nil
}
