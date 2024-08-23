package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

type processHookResult struct {
	distribution     types.Distribution
	votePruned       bool
	vpDiff, newTotal math.Int
}

func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.afterDelegationModified(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship: AfterDelegationModified: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

// afterDelegationModified handles the AfterDelegationModified staking hook. It checks if the delegator has a vote,
// gets the current delegator's voting power gained from the specified validator, gets the x/staking voting power for
// this validator and calls a generic processHook method.
func (h Hooks) afterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	voted, err := h.k.Voted(ctx, delAddr)
	if err != nil {
		return fmt.Errorf("cannot verify if the delegator voted: %w", err)
	}

	// Skip the vote if the delegator doesn't have a vote
	if !voted {
		return nil
	}

	v, found := h.k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return fmt.Errorf("validator not found")
	}

	d, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return fmt.Errorf("delegation not found")
	}

	// Calculate a staking voting power
	stakingVP := v.TokensFromShares(d.GetShares()).Ceil().TruncateInt()

	// Get the current voting power saved in x/sponsorship. If the VP is not found, then we yet don't
	// have a relevant record. This is a valid case when the VP is zero.
	var sponsorshipVP math.Int
	sponsorshipVP, err = h.k.GetDelegatorValidatorPower(ctx, delAddr, valAddr)
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		sponsorshipVP = math.ZeroInt()
	} else if err != nil {
		return fmt.Errorf("cannot get current voting power: %w", err)
	}

	result, err := h.processHook(ctx, delAddr, valAddr, sponsorshipVP, stakingVP)
	if err != nil {
		return fmt.Errorf("cannot process hook: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventVotingPowerUpdate{
		Voter:           delAddr.String(),
		Validator:       valAddr.String(),
		Distribution:    result.distribution,
		VotePruned:      result.votePruned,
		NewVotingPower:  result.newTotal,
		VotingPowerDiff: result.vpDiff,
	})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.beforeDelegationRemoved(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship: BeforeDelegationRemoved: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

// beforeDelegationRemoved handles the BeforeDelegationRemoved staking hook. It checks if the delegator has a vote,
// gets the current delegator's voting power gained from the specified validator, and calls a generic processHook
// method assuming that the x/staking voting power for this validator is zero.
func (h Hooks) beforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	voted, err := h.k.Voted(ctx, delAddr)
	if err != nil {
		return fmt.Errorf("cannot verify if the delegator voted: %w", err)
	}

	// Skip the vote if the delegator doesn't have a vote
	if !voted {
		return nil
	}

	// Get the current voting power saved in x/sponsorship
	// StakingVP is zero in that case, so Diff == (-1) * SponsorshipVP.
	sponsorshipVP, err := h.k.GetDelegatorValidatorPower(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("cannot get current voting power: %w", err)
	}

	result, err := h.processHook(ctx, delAddr, valAddr, sponsorshipVP, math.ZeroInt())
	if err != nil {
		return fmt.Errorf("cannot process hook: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventVotingPowerUpdate{
		Voter:           delAddr.String(),
		Validator:       valAddr.String(),
		Distribution:    result.distribution,
		VotePruned:      result.votePruned,
		NewVotingPower:  result.newTotal,
		VotingPowerDiff: result.vpDiff,
	})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

// processHook is a genetic method to handle changes in delegations. The method:
//  1. Retrieving the vote cast by the delegator
//  2. Calculates the difference between the new (updated) and old (stored in the state) voting power gained from
//     the validator passed as a parameter
//  3. Applies the diff to the total user's voting power
//  4. If the new voting power falls under the minimum required, revoke the vote
//  5. Otherwise, the update the vote, distribution, and voting power records accordingly
//  6. The new voting power might be zero if the user completely undelegated. If it is, the record associated with
//     this validator is deleted.
//
// The method finally returns a struct containing the new distribution, a flag indicating if the vote
// was pruned (revoked), the difference in voting power and the new total voting power.
//
// This function is expected to be used internally by the Hooks type methods.
func (h Hooks) processHook(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	oldVP, newVP math.Int,
) (*processHookResult, error) {
	vote, err := h.k.GetVote(ctx, delAddr)
	if err != nil {
		return nil, fmt.Errorf("could not get vote for delegator '%s': %w", delAddr, err)
	}

	params, err := h.k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get module params: %w", err)
	}

	// Calculate the diff: if it's > 0, then the user has increase it's bond. Otherwise, decreased it's bond.
	powerDiff := newVP.Sub(oldVP)
	newTotalVP := vote.VotingPower.Add(powerDiff)

	// Validate that the user has min voting power. Revoke the vote if not.
	minVP := params.MinVotingPower
	if newTotalVP.LT(minVP) {
		distr, errX := h.k.revokeVote(ctx, delAddr, vote)
		if errX != nil {
			return nil, fmt.Errorf("could not revoke vote: %w", errX)
		}
		return &processHookResult{
			distribution: distr,
			votePruned:   true,
			vpDiff:       powerDiff,
			newTotal:     vote.VotingPower,
		}, nil
	}

	// The code below updates the vote

	// Apply the vote weight breakdown to the diff -> get a distribution update in absolute values
	update := types.ApplyWeights(powerDiff, vote.Weights)

	// Update the current distribution
	distr, err := h.k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return nil, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Update the current user's voting power
	vote.VotingPower = newTotalVP

	// Save the updated vote
	err = h.k.SaveVote(ctx, delAddr, vote)
	if err != nil {
		return nil, fmt.Errorf("cannot save vote: %w", err)
	}

	// Delete the record if the new VP is zero. Otherwise, update the existing.
	if newVP.IsZero() {
		// Delete the voting power cast for this validator
		err = h.k.DeleteDelegatorValidatorPower(ctx, delAddr, valAddr)
		if err != nil {
			return nil, fmt.Errorf("cannot delete delegator validator power: %w", err)
		}
	} else {
		// Update the voting power cast for this validator
		err = h.k.SaveDelegatorValidatorPower(ctx, delAddr, valAddr, newVP)
		if err != nil {
			return nil, fmt.Errorf("cannot save voting power: %w", err)
		}
	}

	return &processHookResult{
		distribution: distr,
		votePruned:   false,
		vpDiff:       powerDiff,
		newTotal:     vote.VotingPower,
	}, nil
}

func (h Hooks) AfterValidatorBeginUnbonding(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error { return nil }

func (h Hooks) BeforeValidatorSlashed(sdk.Context, sdk.ValAddress, sdk.Dec) error { return nil }

func (Hooks) AfterValidatorCreated(sdk.Context, sdk.ValAddress) error { return nil }

func (Hooks) BeforeValidatorModified(sdk.Context, sdk.ValAddress) error { return nil }

func (Hooks) AfterValidatorRemoved(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error { return nil }

func (Hooks) BeforeDelegationCreated(sdk.Context, sdk.AccAddress, sdk.ValAddress) error { return nil }

func (Hooks) AfterUnbondingInitiated(sdk.Context, uint64) error { return nil }

func (Hooks) BeforeDelegationSharesModified(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}
