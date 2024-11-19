package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/invar"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var invs = invar.NamedFuncsList[Keeper]{
	{"delegator-validator-power", InvariantDelegatorValidatorPower},
	{"distribution", InvariantDistribution},
	{"votes", InvariantVotes},
	{"general", InvariantGeneral},
}

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantDelegatorValidatorPower(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		err := k.delegatorValidatorPower.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], value math.Int) (stop bool, err error) {
				if value.IsNegative() {
					return true, fmt.Errorf("negative power: %s", value)
				}
				return false, nil
			})
		if err != nil {
			return err, true
		}
		return nil, false
	}
}

func InvariantDelegatorValidatorPower(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		err := k.delegatorValidatorPower.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], value math.Int) (stop bool, err error) {
				if value.IsNegative() {
					return true, fmt.Errorf("negative power: %s", value)
				}
				return false, nil
			})
		if err != nil {
			return err, true
		}
		return nil, false
	}
}

func InvariantDistribution(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		distribution, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err), true
		}

		// Sum of gauge power <= total power
		totalGaugePower := math.ZeroInt()
		for _, gauge := range distribution.Gauges {
			totalGaugePower = totalGaugePower.Add(gauge.Power)
		}
		if totalGaugePower.GT(distribution.Total) {
			return fmt.Errorf("sum of gauge power %s exceeds total power %s", totalGaugePower, distribution.Total), true
		}

		// All gauge powers non negative
		for _, gauge := range distribution.Gauges {
			if gauge.Power.IsNegative() {
				return fmt.Errorf("negative gauge power: %s", gauge.Power), true
			}
		}

		// Voting power non negative
		if distribution.Total.IsNegative() {
			return fmt.Errorf("negative total voting power: %s", distribution.Total), true
		}

		return nil, false
	}
}

func InvariantDistribution(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		err := k.distribution.Walk(ctx, nil,
			func(value types.Distribution) (stop bool, err error) {
				if value.Total.IsNegative() {
					return true, fmt.Errorf("negative total: %s", value.Total)
				}
				return false, nil
			})
		return nil, false
	}
}

func InvariantVotes(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		// All gauge weights in 1-100
		err := k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			for _, weight := range vote.GaugeWeights {
				if weight < 1 || weight > 100 {
					return true, fmt.Errorf("gauge weight %d out of range (1-100)", weight)
				}
			}
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("iterate votes: %w", err), true
		}

		// Sum of gauge weights equal <= 100
		err = k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			totalWeight := 0
			for _, weight := range vote.GaugeWeights {
				totalWeight += weight
			}
			if totalWeight > 100 {
				return true, fmt.Errorf("sum of gauge weights %d exceeds 100", totalWeight)
			}
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("iterate votes: %w", err), true
		}

		// Voting power non negative
		err = k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			if vote.VotingPower.IsNegative() {
				return true, fmt.Errorf("negative voting power: %s", vote.VotingPower)
			}
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("iterate votes: %w", err), true
		}

		return nil, false
	}
}

func InvariantVotes(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		return nil, false
	}
}

func InvariantGeneral(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		// Total VP of users vote is equal to total VP of records in delegatorsVotingPower
		totalVP := math.ZeroInt()
		err := k.delegatorValidatorPower.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], value math.Int) (stop bool, err error) {
			totalVP = totalVP.Add(value)
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("walk delegator validator power: %w", err), true
		}

		distribution, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err), true
		}

		if !totalVP.Equal(distribution.Total) {
			return fmt.Errorf("total voting power %s does not equal total power in distribution %s", totalVP, distribution.Total), true
		}

		// Distribution is the result of merging the vote.ToDistribution of all votes
		expectedDistribution := types.Distribution{}
		err = k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			expectedDistribution = expectedDistribution.Merge(vote.ToDistribution())
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("iterate votes: %w", err), true
		}

		if !expectedDistribution.Equal(distribution) {
			return fmt.Errorf("distribution does not match expected distribution from votes"), true
		}

		return nil, false
	}
}

func InvariantGeneral(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		return nil, false
	}
}
