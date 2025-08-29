package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "delegator-validator-power", Func: InvariantDelegatorValidatorPower},
	{Name: "distribution", Func: InvariantDistribution},
	{Name: "votes", Func: InvariantVotes},
	{Name: "endorsements", Func: InvariantEndorsements},
	{Name: "general", Func: InvariantGeneral},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func AllInvariantsFn(k Keeper) uinv.Func {
	return invs.AllFn(k)
}

// delegator validator power is non-negative
func InvariantDelegatorValidatorPower(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		err := k.delegatorValidatorPower.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], value math.Int) (stop bool, err error) {
				if value.IsNegative() {
					errs = append(errs, fmt.Errorf("negative power: %s", value))
				}
				return false, nil
			})
		if err != nil {
			return fmt.Errorf("walk delegator validator power: %w", err)
		}
		return errors.Join(errs...)
	})
}

// basic checks on voting power, and consistency with individual gauges
func InvariantDistribution(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		d, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err)
		}
		return d.Validate()
	})
}

// vote weights in range and sum to not more than total (need not be 100% due to abstaining)
func InvariantVotes(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		err := k.IterateVotes(ctx, func(voter sdk.AccAddress, vote types.Vote) (bool, error) {
			errs = append(errs, vote.Validate())
			return false, nil
		})
		errs = append(errs, err)
		return errors.Join(errs...)
	})
}

// endorsement total shares should equal voting power cast to the corresponding RA gauge
func InvariantEndorsements(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		distr, err := k.GetDistribution(ctx)
		if err != nil {
			return fmt.Errorf("get distribution: %w", err)
		}

		for _, gauge := range distr.Gauges {
			g, err := k.incentivesKeeper.GetGaugeByID(ctx, gauge.GaugeId)
			if err != nil {
				return fmt.Errorf("failed to get gauge by id: %d: %w", gauge.GaugeId, err)
			}
			if g.GetRollapp() == nil {
				continue
			}
			endorsement, err := k.GetEndorsement(ctx, g.GetRollapp().RollappId)
			if err != nil {
				return fmt.Errorf("failed to get rollapp by id: %s: %w", g.GetRollapp().RollappId, err)
			}
			if !endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(gauge.Power)) {
				return fmt.Errorf("endorsement total shares does not equal gauge power: rollapp id %s, shares %s: power %s", g.GetRollapp().RollappId, endorsement.TotalShares, gauge.Power)
			}
		}

		return nil
	})
}

// check that across data structures the total voting power and distribution is consistent
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
