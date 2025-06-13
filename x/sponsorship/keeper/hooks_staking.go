package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ stakingtypes.StakingHooks = StakingHooks{}

// StakingHooks wrapper struct for slashing keeper
type StakingHooks struct {
	k Keeper
}

func (k Keeper) StakingHooks() StakingHooks {
	return StakingHooks{k: k}
}

type processHookResult struct {
	distribution     types.Distribution
	votePruned       bool
	vpDiff, newTotal math.Int
}

func (h StakingHooks) AfterDelegationModified(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := h.afterDelegationModified(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship: AfterDelegationModified: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

// afterDelegationModified handles the AfterDelegationModified staking hook. It checks if the delegator has a vote,
// gets the current delegator's voting power gained from the specified validator, gets the x/staking voting power for
// this validator and calls a generic processHook method.
func (h StakingHooks) afterDelegationModified(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	voted, err := h.k.Voted(ctx, delAddr)
	if err != nil {
		return fmt.Errorf("cannot verify if the delegator voted: %w", err)
	}

	// Skip the vote if the delegator doesn't have a vote
	if !voted {
		return nil
	}

	v, err := h.k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("get validator: %w", err)
	}

	d, err := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("get delegation: %w", err)
	}

	// Calculate a staking voting power
	stakingVP := v.TokensFromShares(d.GetShares()).TruncateInt()

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

func (h StakingHooks) BeforeDelegationRemoved(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := h.beforeDelegationRemoved(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship: BeforeDelegationRemoved: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

// beforeDelegationRemoved handles the BeforeDelegationRemoved staking hook. It checks if the delegator has a vote,
// gets the current delegator's voting power gained from the specified validator, and calls a generic processHook
// method assuming that the x/staking voting power for this validator is zero.
func (h StakingHooks) beforeDelegationRemoved(goCtx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
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
// This function is expected to be used internally by the StakingHooks type methods.
func (h StakingHooks) processHook(
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
			newTotal:     newTotalVP,
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

	// Adjust RA endorsement shares with the updated voting power
	err = h.k.UpdateEndorsementsAndPositions(ctx, delAddr, update)
	if err != nil {
		return nil, fmt.Errorf("update endorsements: %w", err)
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

func (h StakingHooks) AfterValidatorBeginUnbonding(context.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (h StakingHooks) AfterValidatorBonded(context.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (h StakingHooks) BeforeValidatorSlashed(context.Context, sdk.ValAddress, math.LegacyDec) error {
	return nil
}

func (StakingHooks) AfterValidatorCreated(context.Context, sdk.ValAddress) error { return nil }

func (StakingHooks) BeforeValidatorModified(context.Context, sdk.ValAddress) error { return nil }

func (StakingHooks) AfterValidatorRemoved(context.Context, sdk.ConsAddress, sdk.ValAddress) error {
	return nil
}

func (StakingHooks) BeforeDelegationCreated(context.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}

func (StakingHooks) AfterUnbondingInitiated(context.Context, uint64) error { return nil }

func (StakingHooks) BeforeDelegationSharesModified(context.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}
