package keeper

import (
	"errors"
	"fmt"
	"strings"

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
	{Name: "staking-sync", Func: InvariantStakingSync},
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

// InvariantStakingSync validates that delegatorValidatorPower matches actual staking delegations
func InvariantStakingSync(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error

		// Check every delegatorValidatorPower entry against actual staking
		err := k.delegatorValidatorPower.Walk(ctx, nil,
			func(key collections.Pair[sdk.AccAddress, sdk.ValAddress], sponsorshipVP math.Int) (stop bool, err error) {
				delAddr := key.K1()
				valAddr := key.K2()

				// Get actual delegation from staking module
				delegation, err := k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
				if err != nil {
					// Check for delegation not found error (the error message pattern)
					if strings.Contains(err.Error(), "no delegation for") {
						// Delegation doesn't exist in staking but exists in sponsorship
						errs = append(errs, fmt.Errorf("delegation not found in staking for delegator %s validator %s, but sponsorship has power %s",
							delAddr, valAddr, sponsorshipVP))
						return false, nil
					}
					return false, fmt.Errorf("get delegation %s -> %s: %w", delAddr, valAddr, err)
				}

				// Get validator to calculate voting power
				validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
				if err != nil {
					return false, fmt.Errorf("get validator %s: %w", valAddr, err)
				}

				// Calculate expected voting power from staking
				expectedVP := validator.TokensFromShares(delegation.GetShares()).TruncateInt()

				// Allow small rounding differences (due to truncation)
				diff := expectedVP.Sub(sponsorshipVP).Abs()
				tolerance := math.NewInt(1) // 1 unit tolerance for rounding

				if diff.GT(tolerance) {
					errs = append(errs, fmt.Errorf("voting power mismatch for delegator %s validator %s: staking=%s sponsorship=%s diff=%s",
						delAddr, valAddr, expectedVP, sponsorshipVP, diff))
				}

				return false, nil
			})
		if err != nil {
			return fmt.Errorf("walk delegator validator power: %w", err)
		}

		return errors.Join(errs...)
	})
}
