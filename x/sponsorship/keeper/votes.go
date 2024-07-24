package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, weights []types.GaugeWeight) (types.Vote, error) {
	params := k.GetParams(ctx)

	// Validate that no gauge got less than MinAllocationWeight and all of them are perpetual
	for _, weight := range weights {
		if weight.Weight.LT(params.MinAllocationWeight) {
			return types.Vote{}, fmt.Errorf("gauge weight '%d' is less than min allocation weight '%d'", weight.Weight.Int64(), params.MinAllocationWeight.Int64())
		}

		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)
		if err != nil {
			return types.Vote{}, fmt.Errorf("failed to get gauge by id '%d': %w", weight.GaugeId, err)
		}

		if !gauge.IsPerpetual {
			return types.Vote{}, fmt.Errorf("gauge '%d' is not perpetual", weight.GaugeId)
		}
	}

	// Get the user’s total voting power from the x/staking
	vpBreakdown, err := k.GetStakingVotingPower(ctx, voter)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to get voting power from x/staking: %w", err)
	}

	// Validate that the user is bonded (power > MinVotingPower)
	if vpBreakdown.TotalPower.LT(params.MinVotingPower) {
		return types.Vote{}, fmt.Errorf("voting power '%d' is less than min voting power expected '%d'", vpBreakdown.TotalPower.Int64(), params.MinVotingPower.Int64())
	}

	vote := types.Vote{
		VotingPower: vpBreakdown.TotalPower,
		Weights:     weights,
	}

	// Apply the weights to the voting power -> now the weights are in absolute values
	update := vote.ToDistribution()

	// Get the current plan from the state
	current, err := k.GetDistribution(ctx)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the user weights to the current plan
	result := current.Merge(update)

	// Save the user's vote
	err = k.SaveVote(ctx, voter, vote)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to save vote: %w", err)
	}

	// Save the user's voting power breakdown
	for _, valPower := range vpBreakdown.Breakdown {
		err = k.SaveVotingPower(ctx, valPower.ValAddr, voter, valPower.Power)
		if err != nil {
			return types.Vote{}, fmt.Errorf("failed to save voting power: %w", err)
		}
	}

	// Save the updated distribution
	err = k.SaveDistribution(ctx, result)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to save distribution: %w", err)
	}

	return vote, nil
}

func (k Keeper) RevokeVote(ctx sdk.Context, voter sdk.AccAddress) error {
	// Get the user’s vote from the state
	vote, err := k.GetVote(ctx, voter)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}

	// Apply the weights to the user’s voting power -> now the weights are in absolute values
	update := vote.ToDistribution()

	// Get the current plan from the state
	current, err := k.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get distribution: %w", err)
	}

	// result = current - update
	result := current.Merge(update.Negate())

	// Save the updated distribution
	err = k.SaveDistribution(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to save distribution: %w", err)
	}

	// Prune the user’s vote and voting power
	k.DeleteVote(ctx, voter)
	k.DeleteVotingPower(ctx, voter) // TODO!

	return nil
}

func (k Keeper) UpdateVotingPower(ctx sdk.Context, voter sdk.AccAddress, power math.Int) error {
	panic("not implemented")
}

type ValidatorPower struct {
	ValAddr sdk.ValAddress
	Power   math.Int
}

type VotingPowerBreakdown struct {
	TotalPower math.Int
	Breakdown  []ValidatorPower
}

// GetStakingVotingPower returns the user's voting power calculated based on the x/staking module.
func (k Keeper) GetStakingVotingPower(ctx sdk.Context, voter sdk.AccAddress) (VotingPowerBreakdown, error) {
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

		if !v.IsBonded() {
			return Continue
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

	return VotingPowerBreakdown{
		TotalPower: totalPower,
		Breakdown:  breakdown,
	}, err
}
