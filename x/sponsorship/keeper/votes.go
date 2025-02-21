package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, weights []types.GaugeWeight) (types.Vote, types.Distribution, error) {
	// Get module params
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("cannot get module params: %w", err)
	}

	// Validate specified weights
	err = k.validateWeights(ctx, weights, params.MinAllocationWeight)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("error validating weights: %w", err)
	}

	// Check if the user's voted. If they have, revoke the previous vote to place a new one.
	voted, err := k.Voted(ctx, voter)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("cannot verify if the voter has already voted: %w", err)
	}
	if voted {
		_, err := k.RevokeVote(ctx, voter)
		if err != nil {
			return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to revoke previous vote: %w", err)
		}
	}

	// Get the user’s total voting power from the x/staking
	vpBreakdown, err := k.GetValidatorBreakdown(ctx, voter)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to get voting power from x/staking: %w", err)
	}

	// Validate that the user has min voting power
	if vpBreakdown.TotalPower.LT(params.MinVotingPower) {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("voting power '%s' is less than min voting power expected '%s'", vpBreakdown.TotalPower, params.MinVotingPower)
	}

	// Apply the vote weights to the power -> get a distribution update in absolute values
	update := types.ApplyWeights(vpBreakdown.TotalPower, weights)

	// Update the current distribution
	distr, err := k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Add voter's shares to RA endorsement shares
	err = k.UpdateEndorsements(ctx, update)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("update endorsements: %w", err)
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

	err = uevent.EmitTypedEvent(ctx, &types.EventVote{
		Voter:        voter.String(),
		Vote:         vote,
		Distribution: distr,
	})
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("emit event: %w", err)
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

	// Subtract voter's shares from RA endorsement shares
	err = k.UpdateEndorsements(ctx, update)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("update endorsements: %w", err)
	}

	// Prune the user’s vote and voting power
	err = k.DeleteVote(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to delete vote: %w", err)
	}
	err = k.DeleteDelegatorPower(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to delete delegator's vote breakdown: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventRevokeVote{
		Voter:        voter.String(),
		Distribution: d,
	})
	if err != nil {
		return types.Distribution{}, fmt.Errorf("emit event: %w", err)
	}

	return d, nil
}

// validateWeights validates that
//   - No gauge get less than MinAllocationWeight
//   - All gauges exist
//   - All gauges are perpetual
func (k Keeper) validateWeights(ctx sdk.Context, weights []types.GaugeWeight, minAllocationWeight math.Int) error {
	for _, weight := range weights {
		// No gauge get less than MinAllocationWeight
		if weight.Weight.LT(minAllocationWeight) {
			return fmt.Errorf("gauge weight is less than min allocation weight: gauge weight %s, min allocation %s", weight.Weight, minAllocationWeight)
		}

		// All gauges exist
		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)
		if err != nil {
			return fmt.Errorf("failed to get gauge by id: %d: %w", weight.GaugeId, err)
		}

		// All gauges are perpetual
		if !gauge.IsPerpetual {
			return fmt.Errorf("gauge is not perpetual: %d", weight.GaugeId)
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
		var valAddr sdk.ValAddress
		valAddr, err = sdk.ValAddressFromBech32(d.GetValidatorAddr())
		if err != nil {
			err = fmt.Errorf("can't convert validator address %s: %w", d.GetValidatorAddr(), err)
			return Break
		}
		var v stakingtypes.Validator
		v, err = k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			err = fmt.Errorf("get validator %s: %w", valAddr, err)
			return Break
		}

		// VotingPower = Ceil(DelegationShares * BondedTokens / TotalShares)
		votingPower := v.TokensFromShares(d.GetShares()).TruncateInt()
		totalPower = totalPower.Add(votingPower)

		breakdown = append(breakdown, ValidatorPower{
			ValAddr: valAddr,
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
