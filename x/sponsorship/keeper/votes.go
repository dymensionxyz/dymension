package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, weights []types.GaugeWeight) (types.Vote, types.Distribution, error) {
	// If the user has already voted, then revoke their vote first
	if k.Voted(ctx, voter) {
		_, err := k.RevokeVote(ctx, voter)
		if err != nil {
			return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to revoke previous vote: %w", err)
		}
	}

	params := k.GetParams(ctx)

	err := k.validateWeights(ctx, weights, params.MinAllocationWeight)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("error validating weights: %w", err)
	}

	// Get the user’s total voting power from the x/staking
	vpBreakdown, err := k.GetValidatorBreakdown(ctx, voter)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to get voting power from x/staking: %w", err)
	}

	// Validate that the user has min voting power
	if vpBreakdown.TotalPower.LT(params.MinVotingPower) {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("voting power '%d' is less than min voting power expected '%d'", vpBreakdown.TotalPower.Int64(), params.MinVotingPower.Int64())
	}

	// Apply the vote weights to the power -> get a distribution update in absolute values
	update := types.ApplyWeights(vpBreakdown.TotalPower, weights)

	// Update the current distribution
	distr, err := k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Save the user's vote
	vote := types.Vote{
		VotingPower: vpBreakdown.TotalPower,
		Weights:     weights,
	}
	err = k.SaveVote(ctx, voter, vote)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to save vote: %w", err)
	}

	// Save the user's voting power breakdown
	for _, valPower := range vpBreakdown.Breakdown {
		err = k.SaveDelegatorValidatorPower(ctx, voter, valPower.ValAddr, valPower.Power)
		if err != nil {
			return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to save voting power: %w", err)
		}
	}

	return vote, distr, nil
}

func (k Keeper) RevokeVote(ctx sdk.Context, voter sdk.AccAddress) (types.Distribution, error) {
	vote, err := k.GetVote(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to get vote: %w", err)
	}
	return k.revokeVote(ctx, voter, vote)
}

// revokeVote revokes a vote by applying the negative user's vote to the current distribution.
// It updates the distribution and prunes the vote and voting power of the voter.
func (k Keeper) revokeVote(ctx sdk.Context, voter sdk.AccAddress, vote types.Vote) (types.Distribution, error) {
	// Apply the weights to the user’s voting power -> now the weights are in absolute values
	update := vote.ToDistribution().Negate()

	// Update the current distribution
	d, err := k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Prune the user’s vote and voting power
	k.DeleteVote(ctx, voter)
	k.DeleteDelegatorPower(ctx, voter)

	return d, nil
}

// validateWeights validates that no gauge got less than MinAllocationWeight and all of them are perpetual
func (k Keeper) validateWeights(ctx sdk.Context, weights []types.GaugeWeight, minAllocationWeight math.Int) error {
	for _, weight := range weights {
		if weight.Weight.LT(minAllocationWeight) {
			return fmt.Errorf("gauge weight '%d' is less than min allocation weight '%d'", weight.Weight.Int64(), minAllocationWeight.Int64())
		}

		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)
		if err != nil {
			return fmt.Errorf("failed to get gauge by id '%d': %w", weight.GaugeId, err)
		}

		if !gauge.IsPerpetual {
			return fmt.Errorf("gauge '%d' is not perpetual", weight.GaugeId)
		}
	}
	return nil
}

type ValidatorPower struct {
	ValAddr sdk.ValAddress // Address of the validator.
	Power   math.Int       // Voting power the user gets from this validator.
}

type ValidatorBreakdown struct {
	TotalPower math.Int // Total power of the breakdown.
	Breakdown  []ValidatorPower
}

// GetValidatorBreakdown returns the user's voting power calculated based on the x/staking module.
func (k Keeper) GetValidatorBreakdown(ctx sdk.Context, voter sdk.AccAddress) (ValidatorBreakdown, error) {
	var err error
	totalPower := math.ZeroInt()
	breakdown := make([]ValidatorPower, 0)

	const Break = true
	const Continue = false

	k.stakingKeeper.IterateDelegatorDelegations(ctx, voter, func(d stakingtypes.Delegation) (stop bool) {
		v, found := k.stakingKeeper.GetValidator(ctx, d.GetValidatorAddr())
		if !found {
			err = fmt.Errorf("can't find validator with address %s", d.GetValidatorAddr())
			return Break
		}

		// VotingPower = Ceil(DelegationShares * BondedTokens / TotalShares)
		votingPower := v.TokensFromShares(d.GetShares()).Ceil().TruncateInt()
		totalPower = totalPower.Add(votingPower)

		breakdown = append(breakdown, ValidatorPower{
			ValAddr: d.GetValidatorAddr(),
			Power:   votingPower,
		})

		return Continue
	})
	if err != nil {
		return ValidatorBreakdown{}, fmt.Errorf("failed to iterate delegator delegations: %w", err)
	}

	return ValidatorBreakdown{
		TotalPower: totalPower,
		Breakdown:  breakdown,
	}, nil
}
