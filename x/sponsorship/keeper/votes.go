package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, weights []types.GaugeWeight) (types.Vote, types.Distribution, error) {
	// Get module params
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("cannot get module params: %w", err)
	}

	// Validate specified weights
	err = k.validateWeights(ctx, weights, params.MinAllocationWeight)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("error validating weights: %w", err)
	}

	// Get the user’s total voting power from the x/staking
	vpBreakdown, err := k.GetValidatorBreakdown(ctx, voter)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to get voting power from x/staking: %w", err)
	}

	// Validate that the user has min voting power
	if vpBreakdown.TotalPower.LT(params.MinVotingPower) {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("voting power '%s' is less than min voting power expected '%s'", vpBreakdown.TotalPower, params.MinVotingPower)
	}

	// Apply the vote weights to the power -> get a distribution update in absolute values
	update := types.ApplyWeights(vpBreakdown.TotalPower, weights)

	// Check if the user's voted. If they have, update the current vote with the existing one.
	voted, err := k.Voted(ctx, voter)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("cannot verify if the voter has already voted: %w", err)
	}
	if voted {
		// Explanation:
		//
		// Let's say we have a global distribution:
		// [1, 1000] [2, 2000] power 3000
		//
		// Imagine a user with 400 DYM staked. He votes 25% on Gauge1 and 75% on Gauge2. When we apply his weights,
		// we convert his vote to a distribution update, which will look like the following:
		// [1, 100] [2, 300] power 400
		//
		// So the distribution after update looks like:
		// [1, 1000] [2, 2000] power 3000 +
		// [1, 100] [2, 300] power 400 =
		// [1, 1100] [2, 2300] power 3400
		//
		// Now imagine this user wants to update his vote (i.e. place a new one while he has the existing).
		// He votes 25% on Gauge1 and 25% on Gauge2, so the distribution update should look like the following:
		// [1, 100] [2, 100] power 400
		//
		// Since the user has the previous vote, we need to account for it as well. For that, we subtract the previous
		// vote from the new one before applying the update (using the merge operation):
		// [1, 100] [2, 100] power 400 -
		// [1, 100] [2, 300] power 400 =
		// [2, -200] power 0 <— Negative power
		//
		// Then, apply this update to the global distribution:
		// [1, 1100] [2, 2300] power 3400 +
		// [2, -200] power 0 =
		// [1, 1100] [2, 2100] power 3400 <— Final distribution
		vote, _ := k.GetVote(ctx, voter)
		// update = newVote - prevVote
		update = update.Merge(vote.ToDistribution().Negate())
	}

	// Update the current distribution
	distr, err := k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Add endorser's shares to RA endorsement shares and update endorser's position
	err = k.UpdateEndorsementsAndPositions(ctx, voter, update)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("update endorsements: %w", err)
	}

	// Save the user's vote
	vote := types.Vote{
		VotingPower: vpBreakdown.TotalPower,
		Weights:     weights,
	}
	err = k.SaveVote(ctx, voter, vote)
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to save vote: %w", err)
	}

	// Save the user's voting power breakdown
	for _, valPower := range vpBreakdown.Breakdown {
		err = k.SaveDelegatorValidatorPower(ctx, voter, valPower.ValAddr, valPower.Power)
		if err != nil {
			return types.Vote{}, types.Distribution{}, fmt.Errorf("failed to save voting power: %w", err)
		}
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventVote{
		Voter:        voter.String(),
		Vote:         vote,
		Distribution: distr,
	})
	if err != nil {
		return types.Vote{}, types.Distribution{}, fmt.Errorf("emit event: %w", err)
	}

	return vote, distr, nil
}

func (k Keeper) RevokeVote(ctx sdk.Context, voter sdk.AccAddress) (types.Distribution, error) {
	vote, err := k.GetVote(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to get vote: %w", err)
	}
	return k.revokeVote(ctx, voter, vote)
}

// revokeVote revokes a vote by applying the negative user's vote to the current distribution.
// It updates the distribution and prunes the vote and voting power of the voter.
func (k Keeper) revokeVote(ctx sdk.Context, voter sdk.AccAddress, vote types.Vote) (types.Distribution, error) {
	// Apply the weights to the user’s voting power -> now the weights are in absolute values
	update := vote.ToDistribution().Negate()

	// Update the current distribution
	d, err := k.UpdateDistribution(ctx, update.Merge)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to update distribution: %w", err)
	}

	// Subtract voter's shares from RA endorsement shares (update is already negated)
	// and update endorser's position
	err = k.UpdateEndorsementsAndPositions(ctx, voter, update)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("update endorsements: %w", err)
	}

	// Prune the user's vote and voting power
	err = k.DeleteVote(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to delete vote: %w", err)
	}
	err = k.DeleteDelegatorPower(ctx, voter)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to delete delegator's vote breakdown: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventRevokeVote{
		Voter:        voter.String(),
		Distribution: d,
	})
	if err != nil {
		return types.Distribution{}, fmt.Errorf("emit event: %w", err)
	}

	return d, nil
}

// validateWeights validates that
//   - No gauge gets less than MinAllocationWeight
//   - All gauges exist
//   - All gauges are perpetual
func (k Keeper) validateWeights(ctx sdk.Context, weights []types.GaugeWeight, minAllocationWeight math.Int) error {
	for _, weight := range weights {
		// No gauge gets less than MinAllocationWeight
		if weight.Weight.LT(minAllocationWeight) {
			return fmt.Errorf("gauge weight is less than min allocation weight: gauge weight %s, min allocation %s", weight.Weight, minAllocationWeight)
		}

		// All gauges exist
		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)
		if err != nil {
			return fmt.Errorf("failed to get gauge by id: %d: %w", weight.GaugeId, err)
		}

		// Voting on endorsement gauges is not supported
		if _, isEndorsement := gauge.DistributeTo.(*incentivestypes.Gauge_Endorsement); isEndorsement {
			return fmt.Errorf("voting on endorsement gauges is not supported: %d", weight.GaugeId)
		}

		// All gauges are perpetual
		if !gauge.IsPerpetual {
			return fmt.Errorf("gauge is not perpetual: %d", weight.GaugeId)
		}
	}
	return nil
}

type ValidatorPower struct {
	ValAddr sdk.ValAddress // Address of the validator.
	Power   math.Int       // Voting power the user gets from this validator.
}

type ValidatorBreakdown struct {
	TotalPower math.Int // Total power of the breakdown.
	Breakdown  []ValidatorPower
}

// GetValidatorBreakdown returns the user's voting power calculated based on the x/staking module.
func (k Keeper) GetValidatorBreakdown(ctx sdk.Context, voter sdk.AccAddress) (ValidatorBreakdown, error) {
	var err, err2 error
	totalPower := math.ZeroInt()
	breakdown := make([]ValidatorPower, 0)

	const Break = true
	const Continue = false

	err2 = k.stakingKeeper.IterateDelegatorDelegations(ctx, voter, func(d stakingtypes.Delegation) (stop bool) {
		var valAddr sdk.ValAddress
		valAddr, err = sdk.ValAddressFromBech32(d.GetValidatorAddr())
		if err != nil {
			err = fmt.Errorf("can't convert validator address %s: %w", d.GetValidatorAddr(), err)
			return Break
		}
		var v stakingtypes.Validator
		v, err = k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			err = fmt.Errorf("get validator %s: %w", valAddr, err)
			return Break
		}

		// VotingPower = Ceil(DelegationShares * BondedTokens / TotalShares)
		votingPower := v.TokensFromShares(d.GetShares()).TruncateInt()
		totalPower = totalPower.Add(votingPower)

		breakdown = append(breakdown, ValidatorPower{
			ValAddr: valAddr,
			Power:   votingPower,
		})

		return Continue
	})
	if err2 != nil {
		return ValidatorBreakdown{}, fmt.Errorf("failed to iterate delegator delegations: %w", err2)
	}
	if err != nil {
		return ValidatorBreakdown{}, fmt.Errorf("failed to get validator breakdown: %w", err)
	}

	return ValidatorBreakdown{
		TotalPower: totalPower,
		Breakdown:  breakdown,
	}, nil
}

// ClearAllVotes efficiently clears all votes and resets the distribution while preserving accumulated rewards.
// This method is equivalent to running RevokeVote on all existing votes but optimized for bulk operations.
// Accumulated rewards in endorsements are preserved by updating endorser positions before clearing their shares.
func (k Keeper) ClearAllVotes(ctx sdk.Context) error {
	// Accumulate rewards for all endorser positions before clearing their voting shares
	err := k.accumulateAllEndorserRewards(ctx)
	if err != nil {
		return fmt.Errorf("failed to accumulate endorser rewards: %w", err)
	}

	// Reset endorsement total shares to zero while preserving accumulated rewards
	err = k.resetEndorsementShares(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset endorsement shares: %w", err)
	}

	err = k.votes.Clear(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}

	err = k.delegatorValidatorPower.Clear(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to clear delegator validator power: %w", err)
	}

	// Reset distribution to empty state
	err = k.SaveDistribution(ctx, types.NewDistribution())
	if err != nil {
		return fmt.Errorf("failed to save empty distribution: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventClearAllVotes{})
	if err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

func (k Keeper) accumulateAllEndorserRewards(ctx sdk.Context) error {
	iterator, err := k.endorserPositions.Iterate(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to iterate endorser positions: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			return fmt.Errorf("failed to get key-value: %w", err)
		}

		voterAddr := kv.Key.K1()
		rollappID := kv.Key.K2()
		endorserPosition := kv.Value

		endorsement, err := k.GetEndorsement(ctx, rollappID)
		if err != nil {
			return fmt.Errorf("failed to get endorsement for rollapp %s: %w", rollappID, err)
		}

		// Update endorser position
		newlyAccruedRewards := endorserPosition.RewardsToBank(endorsement.Accumulator)
		endorserPosition.LastSeenAccumulator = endorsement.Accumulator
		endorserPosition.AccumulatedRewards = endorserPosition.AccumulatedRewards.Add(newlyAccruedRewards...)

		// Reset shares to zero since we're clearing all votes
		endorserPosition.Shares = math.LegacyZeroDec()

		err = k.SaveEndorserPosition(ctx, voterAddr, rollappID, endorserPosition)
		if err != nil {
			return fmt.Errorf("failed to save endorser position for voter %s, rollapp %s: %w", voterAddr, rollappID, err)
		}
	}

	return nil
}

func (k Keeper) resetEndorsementShares(ctx sdk.Context) error {
	iterator, err := k.raEndorsements.Iterate(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to iterate endorsements: %w", err)
	}
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			return fmt.Errorf("failed to get key-value: %w", err)
		}

		endorsement := kv.Value

		endorsement.TotalShares = math.LegacyZeroDec()
		err = k.SaveEndorsement(ctx, endorsement)
		if err != nil {
			return fmt.Errorf("failed to save endorsement for rollapp %s: %w", endorsement.RollappId, err)
		}
	}

	return nil
}
