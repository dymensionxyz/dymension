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
	{Name: "general", Func: InvariantGeneral},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
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

		// Equality check removed https://github.com/dymensionxyz/dymension/blob/ed7dde7e61772051b5595c38da80404d898be6ae/x/sponsorship/keeper/invariants.go#L94-L105

		return nil
	})
}
