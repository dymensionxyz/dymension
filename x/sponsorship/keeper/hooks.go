package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks wrapper struct for slashing keeper
type Hooks struct {
	k Keeper
}

func NewHooks(k Keeper) Hooks {
	return Hooks{k: k}
}

func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.afterDelegationModified(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship hook: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

func (h Hooks) afterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// Get the vote from the state. If it is not present, then the user hasn't voted yet
	vote, err := h.k.GetVote(ctx, delAddr)
	switch {
	case err != nil && errors.Is(err, sdkerrors.ErrNotFound):
		// This user doesn't have a vote
		return nil
	case err != nil:
		return fmt.Errorf("could not get vote: %w", err)
	}

	// Get a validator from x/staking
	v, found := h.k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return fmt.Errorf("validator not found")
	}

	// Get a delegator from x/staking
	d, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if !found {
		return fmt.Errorf("delegation not found")
	}

	// Calculate a staking voting power
	stakingVP := v.TokensFromShares(d.GetShares()).Ceil().TruncateInt()

	// Get the current voting power saved in x/sponsorship
	sponsorshipVP, err := h.k.GetVotingPower(ctx, valAddr, delAddr)
	if err != nil {
		return fmt.Errorf("cannot get current voting power: %w", err)
	}

	// Calculate the diff: if it's > 0, then the user has bonded. Otherwise, unbonded.
	powerDiff := stakingVP.Sub(sponsorshipVP)

	// Apply the vote weight breakdown to the diff -> get a distribution update in absolute values
	update := types.ApplyWeights(powerDiff, vote.Weights)

	// Get the current distribution from the state
	current, err := h.k.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the update for the current distribution
	result := current.Merge(update)

	// Save the updated distribution
	err = h.k.SaveDistribution(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to save distribution: %w", err)
	}

	// Update the current user's voting power and save it back to the state
	vote.VotingPower = vote.VotingPower.Add(powerDiff)
	err = h.k.SaveVote(ctx, delAddr, vote)
	if err != nil {
		return fmt.Errorf("cannot save vote: %w", err)
	}

	return nil
}

func (h Hooks) AfterDelegationModified1(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.processHook(ctx, delAddr, valAddr, func() (oldVP, newVP math.Int, _ error) {
		// Get a validator from x/staking
		v, found := h.k.stakingKeeper.GetValidator(ctx, valAddr)
		if !found {
			return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("validator not found")
		}

		// Get a delegator from x/staking
		d, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
		if !found {
			return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("delegation not found")
		}

		// Calculate a staking voting power
		stakingVP := v.TokensFromShares(d.GetShares()).Ceil().TruncateInt()

		// Get the current voting power saved in x/sponsorship
		sponsorshipVP, err := h.k.GetVotingPower(ctx, valAddr, delAddr)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("cannot get current voting power: %w", err)
		}

		return sponsorshipVP, stakingVP, nil
	})
	if err != nil {
		return fmt.Errorf("sponsorship hook: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.beforeDelegationRemoved(ctx, delAddr, valAddr)
	if err != nil {
		return fmt.Errorf("sponsorship hook: delegator '%s', validator '%s': %w", delAddr, valAddr, err)
	}
	return nil
}

func (h Hooks) beforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	// Get the vote from the state. If it is not present, then the user hasn't voted yet
	vote, err := h.k.GetVote(ctx, delAddr)
	switch {
	case err != nil && errors.Is(err, sdkerrors.ErrNotFound):
		// This user doesn't have a vote
		return nil
	case err != nil:
		return fmt.Errorf("could not get vote for delegator '%'s: %w", delAddr, err)
	}

	// Get the current voting power saved in x/sponsorship
	sponsorshipVP, err := h.k.GetVotingPower(ctx, valAddr, delAddr)
	if err != nil {
		return fmt.Errorf("cannot get current voting power for delegator '%s' and validaror '%s': %w", delAddr, valAddr, err)
	}

	// The diff is stakingVP minus sponsorshipVP. StakingVP is zero in that case, so the diff == (-1) * sponsorshipVP.
	powerDiff := sponsorshipVP.Neg()

	// Apply the vote weight breakdown to the diff -> get a distribution update in absolute values
	update := types.ApplyWeights(powerDiff, vote.Weights)

	// Get the current distribution from the state
	current, err := h.k.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the update for the current distribution
	result := current.Merge(update)

	// Save the updated distribution
	err = h.k.SaveDistribution(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to save distribution: %w", err)
	}

	// Update the current user's voting power and save it back to the state
	vote.VotingPower = vote.VotingPower.Add(powerDiff)
	err = h.k.SaveVote(ctx, delAddr, vote)
	if err != nil {
		return fmt.Errorf("cannot save vote: %w", err)
	}

	// Delete the voting power cast for this validator
	h.k.DeleteVotingPowerForDelegation(ctx, valAddr, delAddr)

	return nil
}

func (h Hooks) BeforeDelegationRemoved1(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	err := h.processHook(ctx, delAddr, valAddr, func() (oldVP, newVP math.Int, _ error) {
		// Get the current voting power saved in x/sponsorship
		sponsorshipVP, err := h.k.GetVotingPower(ctx, valAddr, delAddr)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(),
				fmt.Errorf("cannot get current voting power for delegator '%s' and validaror '%s': %w", delAddr, valAddr, err)
		}

		// The diff is staking VP minus sponsorship VP.
		// Staking VP is zero in that case, so the diff == (-1) * sponsorship VP.
		return sponsorshipVP, math.ZeroInt(), nil
	})
	if err != nil {
		return fmt.Errorf("cannot process hook: %w", err)
	}
	return nil
}

func (h Hooks) processHook(
	ctx sdk.Context,
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	getVotingPowerFn func() (oldVP, newVP math.Int, _ error),
) error {
	// Get the vote from the state. If it is not present, then the user hasn't voted yet
	vote, err := h.k.GetVote(ctx, delAddr)
	switch {
	case err != nil && errors.Is(err, sdkerrors.ErrNotFound):
		// This user doesn't have a vote
		return nil
	case err != nil:
		return fmt.Errorf("could not get vote for delegator '%'s: %w", delAddr, err)
	}

	// Get the currently saved and the updated VP. This method may differ depending on the hook.
	oldVP, newVP, err := getVotingPowerFn()
	if err != nil {
		return fmt.Errorf("fauled to get old and new voting power: %w", err)
	}

	// Calculate the diff: if it's > 0, then the user has bonded. Otherwise, unbonded.
	powerDiff := newVP.Sub(oldVP)

	// Apply the vote weight breakdown to the diff -> get a distribution update in absolute values
	update := types.ApplyWeights(powerDiff, vote.Weights)

	// Get the current distribution from the state
	current, err := h.k.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the update for the current distribution
	result := current.Merge(update)

	// Save the updated distribution
	err = h.k.SaveDistribution(ctx, result)
	if err != nil {
		return fmt.Errorf("failed to save distribution: %w", err)
	}

	// Update the current user's voting power and save it back to the state
	vote.VotingPower = vote.VotingPower.Add(powerDiff)
	err = h.k.SaveVote(ctx, delAddr, vote)
	if err != nil {
		return fmt.Errorf("cannot save vote: %w", err)
	}

	// Delete the record if the new VP is zero. Otherwise, update the existing.
	if newVP.IsZero() {
		// Delete the voting power cast for this validator
		h.k.DeleteVotingPowerForDelegation(ctx, valAddr, delAddr)
	} else {
		// Update the voting power cast for this validator
		err = h.k.SaveVotingPower(ctx, valAddr, delAddr, newVP)
		if err != nil {
			return fmt.Errorf("cannot save voting power: %w", err)
		}
	}

	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	//TODO implement me
	panic("implement me")
}

func (h Hooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	//TODO implement me
	panic("implement me")
}

func (h Hooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) error {
	//TODO implement me
	panic("implement me")
}

func (Hooks) AfterValidatorCreated(sdk.Context, sdk.ValAddress) error { return nil }

func (Hooks) BeforeValidatorModified(sdk.Context, sdk.ValAddress) error { return nil }

func (Hooks) AfterValidatorRemoved(sdk.Context, sdk.ConsAddress, sdk.ValAddress) error { return nil }

func (Hooks) BeforeDelegationCreated(sdk.Context, sdk.AccAddress, sdk.ValAddress) error { return nil }

func (Hooks) AfterUnbondingInitiated(sdk.Context, uint64) error { return nil }

func (Hooks) BeforeDelegationSharesModified(sdk.Context, sdk.AccAddress, sdk.ValAddress) error {
	return nil
}
