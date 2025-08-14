package keeper_test

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestInvariantStakingSync() {
	// Test case 1: Perfect sync - should pass
	s.Run("perfect_sync", func() {
		s.SetupTest()

		// Create validator and delegator
		val := s.CreateValidator()
		valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
		s.Require().NoError(err)

		// Create delegator
		delegationAmount := math.NewInt(1000000)
		bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
		s.Require().NoError(err)

		del := s.CreateDelegator(valAddr, sdk.NewCoin(bondDenom, delegationAmount))
		delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())

		// Create a gauge first
		gaugeID := s.CreateAssetGauge()

		// Create vote to trigger sponsorship tracking
		weights := []types.GaugeWeight{{GaugeId: gaugeID, Weight: types.DYM}} // 1% allocation
		_, _, err = s.App.SponsorshipKeeper.Vote(s.Ctx, delAddr, weights)
		s.Require().NoError(err)

		// Run invariant - should pass
		invariant := keeper.InvariantStakingSync(s.App.SponsorshipKeeper)
		err = invariant(s.Ctx)
		s.Require().NoError(err, "invariant should not be broken with perfect sync")
	})

	// Test case 2: Missing delegation in staking - should fail
	s.Run("missing_delegation_in_staking", func() {
		s.SetupTest()

		// Create validator but no actual delegation
		val := s.CreateValidator()
		valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
		s.Require().NoError(err)

		// Create a test account
		testAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, types.DYM.MulRaw(1_000))
		delAddr := testAddrs[0]

		// Manually inject a sponsorship delegatorValidatorPower entry without a corresponding staking delegation
		// This simulates the corruption scenario where staking data was deleted but sponsorship wasn't updated
		corruptedPower := math.NewInt(1000000)
		err = s.App.SponsorshipKeeper.SaveDelegatorValidatorPower(s.Ctx, delAddr, valAddr, corruptedPower)
		s.Require().NoError(err)

		// Verify the corruption exists
		hasPower, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, delAddr, valAddr)
		s.Require().NoError(err)
		s.Require().True(hasPower, "sponsorship should have power entry")

		// Verify no delegation exists in staking
		_, stakingErr := s.App.StakingKeeper.GetDelegation(s.Ctx, delAddr, valAddr)
		s.Require().Error(stakingErr, "staking should not have delegation")

		// Run invariant - should fail
		invariant := keeper.InvariantStakingSync(s.App.SponsorshipKeeper)
		err = invariant(s.Ctx)
		s.Require().Error(err, "invariant should be broken when delegation is missing in staking")
		s.Require().True(errorsmod.IsOf(err, uinv.ErrBroken), "error should be marked as invariant breaking")
	})

	// Test case 3: Voting power mismatch - should fail
	s.Run("voting_power_mismatch", func() {
		s.SetupTest()

		// Create validator and delegator
		val := s.CreateValidator()
		valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
		s.Require().NoError(err)

		// Create delegator
		delegationAmount := math.NewInt(1000000)
		bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
		s.Require().NoError(err)

		del := s.CreateDelegator(valAddr, sdk.NewCoin(bondDenom, delegationAmount))
		delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())

		// Create vote
		// Create a gauge first
		gaugeID := s.CreateAssetGauge()
		weights := []types.GaugeWeight{{GaugeId: gaugeID, Weight: types.DYM}}
		_, _, err = s.App.SponsorshipKeeper.Vote(s.Ctx, delAddr, weights)
		s.Require().NoError(err)

		// Manually corrupt the delegatorValidatorPower (simulate state corruption)
		corruptedPower := math.NewInt(2000000) // Double the actual power
		err = s.App.SponsorshipKeeper.SaveDelegatorValidatorPower(s.Ctx, delAddr, valAddr, corruptedPower)
		s.Require().NoError(err)

		// Run invariant - should fail
		invariant := keeper.InvariantStakingSync(s.App.SponsorshipKeeper)
		err = invariant(s.Ctx)
		s.Require().Error(err, "invariant should be broken when voting power is mismatched")
		s.Require().True(errorsmod.IsOf(err, uinv.ErrBroken), "error should be marked as invariant breaking")
	})

	// Test case 4: No voting power entries - should pass
	s.Run("no_voting_power_entries", func() {
		s.SetupTest()

		// Run invariant with no delegatorValidatorPower entries - should pass
		invariant := keeper.InvariantStakingSync(s.App.SponsorshipKeeper)
		err := invariant(s.Ctx)
		s.Require().NoError(err, "invariant should not be broken with no entries")
	})

	// Test case 5: Multiple delegations - some correct, some incorrect
	s.Run("mixed_delegations", func() {
		s.SetupTest()

		// Create validators
		val1 := s.CreateValidator()
		valAddr1, err := sdk.ValAddressFromBech32(val1.GetOperator())
		s.Require().NoError(err)

		val2 := s.CreateValidator()
		valAddr2, err := sdk.ValAddressFromBech32(val2.GetOperator())
		s.Require().NoError(err)

		// Create delegators
		delegationAmount := math.NewInt(1000000)
		bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
		s.Require().NoError(err)

		del1 := s.CreateDelegator(valAddr1, sdk.NewCoin(bondDenom, delegationAmount))
		delAddr1 := sdk.MustAccAddressFromBech32(del1.GetDelegatorAddr())

		del2 := s.CreateDelegator(valAddr2, sdk.NewCoin(bondDenom, delegationAmount))
		delAddr2 := sdk.MustAccAddressFromBech32(del2.GetDelegatorAddr())

		// Create votes for both
		// Create a gauge first
		gaugeID := s.CreateAssetGauge()
		weights := []types.GaugeWeight{{GaugeId: gaugeID, Weight: types.DYM}}
		_, _, err = s.App.SponsorshipKeeper.Vote(s.Ctx, delAddr1, weights)
		s.Require().NoError(err)
		_, _, err = s.App.SponsorshipKeeper.Vote(s.Ctx, delAddr2, weights)
		s.Require().NoError(err)

		// Manually inject a corrupted sponsorship entry for a non-existent delegation
		// This simulates corruption where sponsorship has power for a delegation that doesn't exist
		testAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, types.DYM.MulRaw(1_000))
		corruptedDelAddr := testAddrs[0]
		corruptedPower := math.NewInt(5000000)
		err = s.App.SponsorshipKeeper.SaveDelegatorValidatorPower(s.Ctx, corruptedDelAddr, valAddr1, corruptedPower)
		s.Require().NoError(err)

		// Run invariant - should fail because of the corrupted delegation
		invariant := keeper.InvariantStakingSync(s.App.SponsorshipKeeper)
		err = invariant(s.Ctx)
		s.Require().Error(err, "invariant should be broken when one delegation is corrupted")
		s.Require().True(errorsmod.IsOf(err, uinv.ErrBroken), "error should be marked as invariant breaking")
	})
}

// Test all invariants together
func (s *KeeperTestSuite) TestAllInvariants() {
	s.SetupTest()

	// Create validator and delegator
	val := s.CreateValidator()
	valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
	s.Require().NoError(err)

	// Create delegator
	delegationAmount := math.NewInt(1000000)
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	del := s.CreateDelegator(valAddr, sdk.NewCoin(bondDenom, delegationAmount))
	delAddr := sdk.MustAccAddressFromBech32(del.GetDelegatorAddr())

	// Create a gauge first
	gaugeID := s.CreateAssetGauge()
	weights := []types.GaugeWeight{{GaugeId: gaugeID, Weight: types.DYM}}
	_, _, err = s.App.SponsorshipKeeper.Vote(s.Ctx, delAddr, weights)
	s.Require().NoError(err)

	// Test that all invariants pass (AllInvariants returns (string, bool))
	allInvariants := keeper.AllInvariants(s.App.SponsorshipKeeper)
	msg, broken := allInvariants(s.Ctx)
	s.Require().False(broken, "all invariants should pass with normal voting scenario. Message: %s", msg)
}
