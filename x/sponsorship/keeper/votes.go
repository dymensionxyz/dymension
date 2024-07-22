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
	votingPower, err := k.GetStakingVotingPower(ctx, voter)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to get voting power from x/staking: %w", err)
	}

	// Validate that the user is bonded (power > MinVotingPower)
	if votingPower.LT(params.MinVotingPower) {
		return types.Vote{}, fmt.Errorf("voting power '%d' is less than min voting power expected '%d'", votingPower.Int64(), params.MinVotingPower.Int64())
	}

	vote := types.Vote{
		VotingPower: votingPower,
		Weights:     weights,
	}

	// Apply the weights to the voting power -> now the weights are in absolute values
	distrUpdate := vote.ToDistribution()

	// Get the current plan from the state
	distr, err := k.GetDistribution(ctx)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to get distribution from the state: %w", err)
	}

	// Apply the user weights to the current plan and save it back to the state
	resultDistr := types.ApplyUpdate(distr, distrUpdate)
	err := k.SaveDistribution(ctx, resultDistr)
	if err != nil {
		return types.Vote{}, fmt.Errorf("failed to save distribution: %w", err)
	}

	return vote, nil
}

func (k Keeper) RevokeVote(ctx sdk.Context, voter sdk.AccAddress) error {
	panic("not implemented")
}

func (k Keeper) UpdateVotingPower(ctx sdk.Context, voter sdk.AccAddress, power math.Int) error {
	panic("not implemented")
}

// GetStakingVotingPower returns the user's voting power calculated based on the x/staking module.
func (k Keeper) GetStakingVotingPower(ctx sdk.Context, voter sdk.AccAddress) (math.Int, error) {
	var err error
	totalPower := math.ZeroInt()

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
		return Continue
	})

	return totalPower, err
}
