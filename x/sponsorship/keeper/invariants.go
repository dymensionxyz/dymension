package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "delegator-validator-power", Func: InvariantDelegatorValidatorPower},
	{Name: "distribution", Func: InvariantDistribution},
	{Name: "votes", Func: InvariantVotes},
	{Name: "general", Func: InvariantGeneral},
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
		return k.delegatorValidatorPower.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], value math.Int) (stop bool, err error) {
				if value.IsNegative() {
					return false, fmt.Errorf("negative power: %s", value)
				}
				return false, nil
			})
	})
}

func InvariantDistribution(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		d, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err)
		}

		t := math.ZeroInt()
		for _, g := range d.Gauges {
			t = t.Add(g.Power)
		}
		if t.GT(d.VotingPower) {
			return fmt.Errorf("sum of gauge power exceeds total power sum: %s: total: %s", t, d.VotingPower)
		}

		var errs []error
		for _, g := range d.Gauges {
			if g.Power.IsNegative() {
				errs = append(errs, fmt.Errorf("negative g power: %d: %s", g.GaugeId, g.Power))
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		if d.VotingPower.IsNegative() {
			return fmt.Errorf("negative total voting power: %s", d.VotingPower)
		}

		return nil
	})
}

func InvariantVotes(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		err := k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			if vote.VotingPower.IsNegative() {
				return false, fmt.Errorf("negative voting power: %s", vote.VotingPower)
			}
			t := sdk.ZeroInt()
			for _, weight := range vote.GetWeights() {
				w := weight.Weight
				if w.LT(sdk.OneInt()) || w.GT(sdk.NewInt(100)) {
					return false, fmt.Errorf("gauge weight out of range (1-100): %s", w)
				}
				t = t.Add(w)
			}
			if t.GT(sdk.NewInt(100)) {
				return false, fmt.Errorf("sum of gauge weights exceeds 100: %s", t)
			}
			return false, nil
		})

		return errorsmod.Wrap(err, "iterate votes")
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
			return fmt.Errorf("sum delegator validator power: %w", err)
		}

		distribution, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err)
		}

		if !totalVP.Equal(distribution.VotingPower) {
			return fmt.Errorf("total voting power does not equal total power in distribution: total: %s: distr:  %s", totalVP, distribution.VotingPower)
		}

		expectedDistribution := types.NewDistribution()
		err = k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error) {
			expectedDistribution = expectedDistribution.Merge(vote.ToDistribution())
			return false, nil
		})
		if err != nil {
			return fmt.Errorf("merge votes: %w", err)
		}

		if !expectedDistribution.Equal(distribution) {
			return fmt.Errorf("distribution does not match expected distribution from votes")
		}

		return nil
	})
}
