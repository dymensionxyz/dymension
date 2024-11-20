package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{"delegator-validator-power", InvariantDelegatorValidatorPower},
	{"distribution", InvariantDistribution},
	{"votes", InvariantVotes},
	{"general", InvariantGeneral},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantDelegatorValidatorPower(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {

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
	})
}

func InvariantDistribution(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {

		d, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err), true
		}

		t := math.ZeroInt()
		for _, g := range d.Gauges {
			t = t.Add(g.Power)
		}
		if t.GT(d.VotingPower) {
			return fmt.Errorf("sum of gauge power exceeds total power sum: %s: total: %s", t, d.VotingPower), true
		}

		for _, g := range d.Gauges {
			if g.Power.IsNegative() {
				return fmt.Errorf("negative g power: %s", g.Power), true
			}
		}

		if d.VotingPower.IsNegative() {
			return fmt.Errorf("negative total voting power: %s", d.VotingPower), true
		}

		return nil, false
	})
}

func InvariantVotes(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {

		// All gauge weights in 1-100
		err := k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			if vote.VotingPower.IsNegative() {
				return true, fmt.Errorf("negative voting power: %s", vote.VotingPower)
			}
			t := sdk.ZeroInt()
			for _, weight := range vote.GetWeights() {
				w := weight.Weight
				if w.LT(sdk.OneInt()) || w.GT(sdk.NewInt(100)) {
					return true, fmt.Errorf("gauge weight out of range (1-100): %s", w)
				}
				t = t.Add(w)
			}
			if t.GT(sdk.NewInt(100)) {
				return true, fmt.Errorf("sum of gauge weights exceeds 100: %s", t)
			}
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("iterate votes: %w", err), true
		}
		return nil, false
	})
}

func InvariantGeneral(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {

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

		if !totalVP.Equal(distribution.VotingPower) {
			return fmt.Errorf("total voting power does not equal total power in distribution: total: %s: distr:  %s", totalVP, distribution.VotingPower), true
		}

		expectedDistribution := types.NewDistribution()
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
	})
}
