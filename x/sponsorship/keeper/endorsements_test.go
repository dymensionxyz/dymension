package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestEndorsements() {
	// Start the initial epoch
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Create a rollapp gauge (ID 1)
	rollappID := s.CreateDefaultRollapp()
	const rollappGaugeID = 1

	// Create an endorsement gauge (ID 2)
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	dym1000 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000)))
	s.FundAcc(gaugeCreator, dym1000)

	// Gauge for 1000 DYM and 10 epochs
	endorsementGaugeID, err := s.App.IncentivesKeeper.CreateEndorsementGauge(
		s.Ctx,
		false,
		gaugeCreator,
		dym1000,
		incentivestypes.EndorsementGauge{
			RollappId: rollappID,
		},
		s.Ctx.BlockTime(),
		10,
	)
	s.Require().NoError(err)

	// Create validators and delegators
	val := s.CreateValidator()
	valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
	s.Require().NoError(err)

	// User1 delegates 40 DYM (total voting power)
	initial1 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))
	del1 := s.CreateDelegator(valAddr, initial1)

	// User2 delegates 100 DYM (but only endorses with 60 shares in step 3)
	initial2 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))
	del2 := s.CreateDelegator(valAddr, initial2)

	// Helper to create 100 DYM coins
	dym0 := sdk.NewCoins()
	dym60 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(60)))
	dym100 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	dym160 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(160)))
	dym200 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(200)))
	dym260 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(260)))
	dym360 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(360)))

	/***************************************************************/
	/*                           Epoch 0                           */
	/***************************************************************/

	// Test if we try to distribute rewards to the empty endorsement, then nothing will happen
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	eg, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(eg.DistributedCoins.IsZero())
	s.Require().Equal(uint64(0), eg.FilledEpochs)

	// Verify initial state: accumulator = 0.0, total shares = 0
	endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.IsZero(), "accumulator should be 0.0")
	s.Require().True(endorsement.TotalShares.IsZero(), "total shares should be 40")
	s.Require().True(endorsement.TotalCoins.IsZero())
	s.Require().True(endorsement.DistributedCoins.IsZero())

	/***************************************************************/
	/*                           Epoch 1                           */
	/***************************************************************/

	// User1 endorses with 40 shares
	// User1 votes 40% (of 100-DYM delegation) on the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del1.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(40)}, // 40% to rollapp gauge
		},
	})

	// Verify initial state: accumulator = 0.0, total shares = 40
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.IsZero(), "accumulator should be 0.0")
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(40))), "total shares should be 40")
	unlockedCoins := endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym0), "unlocked coins mismatch: expected %s, got %s", dym0, unlockedCoins)

	// User1 should have 0 claimable rewards (nothing distributed yet)
	user1Addr := sdk.MustAccAddressFromBech32(del1.GetDelegatorAddr())
	result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable rewards initially")

	/***************************************************************/
	/*                           Epoch 2                           */
	/***************************************************************/

	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify accumulator = 2.5 (100 DYM / 40 shares)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulator := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(25, 1))) // 2.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should be 2.5")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym100), "unlocked coins mismatch: expected %s, got %s", dym100, unlockedCoins)

	// User1 should have 100 DYM claimable: (2.5 - 0) * 40 = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100), "user1 should have 100 DYM claimable")

	// User2 endorses with 60 shares
	// User2 votes 60% on the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(60)}, // 60% to rollapp gauge
		},
	})

	// Verify total shares = 100 (40 + 60)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(100))), "total shares should be 100")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym100), "unlocked coins mismatch: expected %s, got %s", dym100, unlockedCoins)

	// User2 should have 0 claimable (LSA = 2.5, so (2.5 - 2.5) * 60 = 0)
	user2Addr := sdk.MustAccAddressFromBech32(del2.GetDelegatorAddr())
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user2 should have 0 claimable rewards initially")

	// User1 should still have 100 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100), "user1 should still have 100 DYM claimable")

	/***************************************************************/
	/*                           Epoch 3                           */
	/***************************************************************/

	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify accumulator = 3.5 (2.5 + 100/100)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulator = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(35, 1))) // 3.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should be 3.5")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym200), "unlocked coins mismatch: expected %s, got %s", dym200, unlockedCoins)

	// User1 should have 140 DYM claimable: (3.5 - 0) * 40 = 140
	dym140 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(140)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym140), "user1 should have 140 DYM claimable")

	// User2 should have 60 DYM claimable: (3.5 - 2.5) * 60 = 60
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should have 60 DYM claimable")

	// Claim User1 and verify their balance
	{
		// Snapshot User1 balance before claiming
		beforeBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 claims
		err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, rollappGaugeID)
		s.Require().NoError(err)

		// Snapshot User1 balance after claiming
		afterBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 should have increased by 140 DYM
		s.Require().True(afterBalance.Sub(beforeBalance).Amount.Equal(dym140[0].Amount), "user1 balance should have increased by 140 DYM")
	}

	// User1 should now have 0 claimable (LSA updated to 3.5)
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable after claiming")

	// User2 should still have 60 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should still have 60 DYM claimable")

	// User2 un-endorses
	s.RevokeVote(types.MsgRevokeVote{
		Voter: user2Addr.String(),
	})

	// User2 should still have 60 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should still have 60 DYM claimable")

	// Verify total shares = 40 (only user1 left)
	// After User1 claimed 140 DYM, and User2 un-endorsed. Unlocked coins should be 60.
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(40))), "total shares should be 40 after user2 leaves")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym60), "unlocked coins mismatch: expected %s, got %s", dym60, unlockedCoins)

	/***************************************************************/
	/*                           Epoch 4                           */
	/***************************************************************/

	// +100 DYM
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify accumulator = 6.0 (3.5 + 100/40)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulator = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(6))) // 6.0
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should be 6.0")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym160), "unlocked coins mismatch: expected %s, got %s", dym160, unlockedCoins)

	// User1 should have 100 DYM claimable: (6 - 3.5) * 40 = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100), "user1 should have 100 DYM claimable")

	// User1 un-endorses
	s.RevokeVote(types.MsgRevokeVote{
		Voter: del1.GetDelegatorAddr(),
	})

	// Claim User1 and verify their balance
	{
		// Snapshot User1 balance before claiming
		beforeBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 claims
		err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, rollappGaugeID)
		s.Require().NoError(err)

		// Snapshot User1 balance after claiming
		afterBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 should have increased by 100 DYM
		s.Require().True(afterBalance.Sub(beforeBalance).Amount.Equal(dym100[0].Amount), "user1 balance should have increased by 100 DYM")
	}

	// Verify total shares = 0 (no endorsers left)
	// After User1 claimed 100 DYM. Unlocked coins should be 60 (160 - 100).
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.IsZero(), "total shares should be 0 after all users leave")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym60), "unlocked coins mismatch: expected %s, got %s", dym60, unlockedCoins)

	// User2 re-endorses with 100 shares
	// User2 now votes with 100 DYM equivalent
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// User2 should have 60 DYM claimable bc of accumulated rewards
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should have 60 DYM claimable")

	/***************************************************************/
	/*                           Epoch 5                           */
	/***************************************************************/

	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Get current total shares to calculate expected accumulator
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)

	// Expected accumulator = 6.0 + (100 / 100)
	expectedAccumulatorValue := math.LegacyNewDec(6).Add(math.LegacyNewDec(100).Quo(math.LegacyNewDec(100)))
	expectedAccumulator = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, expectedAccumulatorValue))

	// Check endorsement
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(100))), "total shares should be 100 on epoch 5")
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "global accumulator should be 7.0 on epoch 5")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym160), "unlocked coins mismatch: expected %s, got %s", dym160, unlockedCoins)

	// User2 should have 160 DYM claimable: (new_accumulator - 6.0) * shares + accumulated = 160
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100.Add(dym60...)), "user2 should have 160 DYM claimable after final distribution")

	/***************************************************************/
	/*                       Staking Events                        */
	/***************************************************************/

	// User2 stakes 100 DYM more
	s.Delegate(user2Addr, valAddr, dym200[0])

	// Verify accumulator is still the same – 7.0, but total shares are increased
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulator = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(7))) // 7.0
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should be 7.0")
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(300))), "total shares should be 300 after staking")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym160), "unlocked coins mismatch: expected %s, got %s", dym160, unlockedCoins)

	// Verify AUR becomes 160 DYM, LSA becomes 7.0, shares become 300
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	expectedAUR := dym100.Add(dym60...) // 160 DYM
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator)
	s.Require().Truef(pos.Shares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(300))), "Shares mismatch. Expected %s, got %s", math.LegacyNewDecFromInt(types.DYM.MulRaw(300)), pos.Shares)

	// User2 should have 160 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedAUR), "user2 should have 160 DYM claimable after staking events")

	/***************************************************************/
	/*                  User2 unstakes 100 DYM                     */
	/***************************************************************/
	// User2 currently has 300 DYM staked. Unstaking 100 DYM.
	s.Undelegate(user2Addr, valAddr, sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))

	// Verify Endorsement: GA should be 7.0, TotalShares should be 200
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "GA should still be 7.0") // expectedAccumulator is GA 7.0
	dym200SharesDec := math.LegacyNewDecFromInt(types.DYM.MulRaw(200))
	s.Require().True(endorsement.TotalShares.Equal(dym200SharesDec), "TotalShares should be 200 after unstake")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym160), "unlocked coins mismatch: expected %s, got %s", dym160, unlockedCoins)

	// Verify User2 Position: AUR=160, LSA=7.0, Shares=200
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator)
	s.Require().Truef(pos.Shares.Equal(dym200SharesDec), "Shares mismatch. Expected %s, got %s", dym200SharesDec, pos.Shares)

	// User2 should still have 160 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedAUR), "User2 claimable should be 160 DYM")

	/***************************************************************/
	/*                           Epoch 6                           */
	/***************************************************************/
	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify Endorsement: GA = 7.0 + (100/200) = 7.5. TotalShares = 200.
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulatorEpoch6 := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(75, 1))) // 7.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorEpoch6), "GA should be 7.5 in Epoch 6")
	s.Require().True(endorsement.TotalShares.Equal(dym200SharesDec), "TotalShares should be 200 in Epoch 6")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym260), "unlocked coins mismatch: expected %s, got %s", dym260, unlockedCoins)

	// Verify User2 Position (unchanged by epoch start): AUR=160, LSA=7.0, Shares=200
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator) // LSA is still 7.0
	s.Require().Truef(pos.Shares.Equal(dym200SharesDec), "Shares mismatch. Expected %s, got %s", dym200SharesDec, pos.Shares)

	// User2 Claimable: (7.5 - 7.0) * 200 + 160 = 100 + 160 = 260 DYM
	expectedClaimableEpoch6 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(260)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch6), "User2 claimable should be 260 DYM in Epoch 6")

	/***************************************************************/
	/*               User2 removes 100 shares (updates vote)       */
	/***************************************************************/
	// User2 (total stake 200) updates vote from 200 shares to 100 shares
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(), // user2Addr.String()
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(50)}, // 50% of 200 DYM -> 100 shares
		},
	})

	// Verify Endorsement: GA=7.5, TotalShares = 100 (200(old total) - 200(old U2) + 100(new U2))
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorEpoch6), "GA should be 7.5 after vote update")
	dym100SharesDec := math.LegacyNewDecFromInt(types.DYM.MulRaw(100))
	s.Require().True(endorsement.TotalShares.Equal(dym100SharesDec), "TotalShares should be 100 after vote update")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym260), "unlocked coins mismatch: expected %s, got %s", dym260, unlockedCoins)

	// Verify User2 Position:
	// Rewards to bank: (7.5 - 7) * 200 = 100. New AUR = 160 + 100 = 260. LSA = 7.5. Shares = 100.
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch6.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch6, pos.AccumulatedRewards)       // AUR is 260
	s.Require().Truef(expectedAccumulatorEpoch6.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulatorEpoch6, pos.LastSeenAccumulator) // LSA is 7.5
	s.Require().Truef(pos.Shares.Equal(dym100SharesDec), "Shares mismatch. Expected %s, got %s", dym100SharesDec, pos.Shares)

	// User2 Claimable: (7.5 - 7.5) * 100 + 260 = 260 DYM
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch6), "User2 claimable should be 260 DYM after vote update")

	/***************************************************************/
	/*                           Epoch 7                           */
	/***************************************************************/
	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify Endorsement: GA = 7.5 + (100/100) = 8.5. TotalShares = 100.
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulatorEpoch7 := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(85, 1))) // 8.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorEpoch7), "GA should be 8.5 in Epoch 7")
	s.Require().True(endorsement.TotalShares.Equal(dym100SharesDec), "TotalShares should be 100 in Epoch 7")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym360), "unlocked coins mismatch: expected %s, got %s", dym360, unlockedCoins)

	// Verify User2 Position (AUR and LSA unchanged by epoch start): AUR=260, LSA=7.5, Shares=100
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch6.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch6, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulatorEpoch6.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulatorEpoch6, pos.LastSeenAccumulator) // LSA is still 7.5
	s.Require().Truef(pos.Shares.Equal(dym100SharesDec), "Shares mismatch. Expected %s, got %s", dym100SharesDec, pos.Shares)

	// User2 Claimable: (8.5 - 7.5) * 100 + 260 = 100 + 260 = 360 DYM
	expectedClaimableEpoch7 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(360)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch7), "User2 claimable should be 360 DYM in Epoch 7")

	/***************************************************************/
	/*            User2 unstakes all (vote is removed)             */
	/***************************************************************/
	// User2 currently has 200 DYM staked. Unstaking all 200 DYM.
	s.Undelegate(user2Addr, valAddr, sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(200)))

	// Verify Endorsement: GA=8.5, TotalShares = 0
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorEpoch7), "GA should be 8.5 after unstake all")
	s.Require().True(endorsement.TotalShares.IsZero(), "TotalShares should be 0 after unstake all")
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym360), "unlocked coins mismatch: expected %s, got %s", dym360, unlockedCoins)

	// Verify User2 Position:
	// Rewards to bank: (8.5 - 7.5) * 100 = 100. New AUR = 260 + 100 = 360. LSA = 8.5. Shares = 0.
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch7.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch7, pos.AccumulatedRewards)       // AUR is 360
	s.Require().Truef(expectedAccumulatorEpoch7.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulatorEpoch7, pos.LastSeenAccumulator) // LSA is 8.5
	s.Require().True(pos.Shares.IsZero(), "Shares mismatch. Expected 0, got %s", pos.Shares)

	// User2 Claimable: (8.5 - 8.5) * 0 + 360 = 360 DYM
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch7), "User2 claimable should be 360 DYM after unstake all")

	user2Voted, err := s.App.SponsorshipKeeper.Voted(s.Ctx, user2Addr)
	s.Require().NoError(err)
	s.Require().False(user2Voted, "User2 should not have voted after unstake all")

	/***************************************************************/
	/*                         User2 claims                        */
	/***************************************************************/
	// Claim User2 and verify the balance change
	{
		balanceBeforeClaim := s.App.BankKeeper.GetBalance(s.Ctx, user2Addr, sdk.DefaultBondDenom)

		err = s.App.SponsorshipKeeper.Claim(s.Ctx, user2Addr, rollappGaugeID)
		s.Require().NoError(err)

		balanceAfterClaim := s.App.BankKeeper.GetBalance(s.Ctx, user2Addr, sdk.DefaultBondDenom)

		s.Require().True(balanceAfterClaim.Sub(balanceBeforeClaim).Amount.Equal(expectedClaimableEpoch7[0].Amount), "User2 claimed amount mismatch")
	}

	// Verify User2 Position: AUR=0, LSA=8.5, Shares=0
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero after claim. Got %s", pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulatorEpoch7.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulatorEpoch7, pos.LastSeenAccumulator)
	s.Require().True(pos.Shares.IsZero(), "Shares should be zero after claim. Got %s", pos.Shares)

	// User2 Claimable should be 0
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "User2 claimable should be zero after claim")

	// Verify Unlocked Coins after User2's final claim
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...)
	s.Require().True(unlockedCoins.Equal(dym0), "unlocked coins mismatch: expected %s, got %s after final claim", dym0, unlockedCoins)

	/***************************************************************/
	/*     Scenario: user has share with repeating decimal       */
	/***************************************************************/
	s.T().Log("Scenario: user has share with repeating decimal")

	// User1 (del1Addr) currently has 100 DYM staked.
	// To make User1 contribute exactly 60 shares (60M units) with a 100% vote,
	// their stake needs to be 60 DYM. Undelegate 40 DYM.
	s.Undelegate(user1Addr, valAddr, sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40)))
	s.Require().True(s.App.StakingKeeper.Validator(s.Ctx, valAddr).GetDelegatorShares().Equal(math.LegacyNewDecFromInt(commontypes.DYM.MulRaw(60))), "Validator shares should be 60M after User1 partial undelegation")

	// User1 endorses with 60 shares (by voting 100% of their 60 DYM stake)
	s.Vote(types.MsgVote{
		Voter: user1Addr.String(),
		Weights: []types.GaugeWeight{
			// Weight of 1 (or any positive number) means 100% to this gauge if it's the only one
			{GaugeId: rollappGaugeID, Weight: math.NewInt(1)},
		},
	})

	// Verify Endorsement: GA=8.5 (from previous scenario), TotalShares=60M
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	// expectedAccumulatorEpoch7 was 8.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorEpoch7), "GA should be 8.5 before new epoch. Expected %s, got %s", expectedAccumulatorEpoch7, endorsement.Accumulator)
	expectedUser1SharesDec := math.LegacyNewDecFromInt(commontypes.DYM.MulRaw(60))
	s.Require().True(endorsement.TotalShares.Equal(expectedUser1SharesDec), "TotalShares should be 60M. Expected %s, got %s", expectedUser1SharesDec, endorsement.TotalShares)
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...) // Should still be 0 from previous claims
	s.Require().True(unlockedCoins.IsZero(), "Unlocked coins should be 0 before new epoch. Got %s", unlockedCoins)

	// Verify User1 Position: Shares=60M, LSA=8.5, AUR=0
	posUser1, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user1Addr, rollappID)
	s.Require().NoError(err)
	s.Require().True(posUser1.Shares.Equal(expectedUser1SharesDec), "User1 shares should be 60M. Expected %s, got %s", expectedUser1SharesDec, posUser1.Shares)
	s.Require().True(posUser1.LastSeenAccumulator.Equal(expectedAccumulatorEpoch7), "User1 LSA should be 8.5. Expected %s, got %s", expectedAccumulatorEpoch7, posUser1.LastSeenAccumulator)
	s.Require().True(posUser1.AccumulatedRewards.IsZero(), "User1 AUR should be 0. Got %s", posUser1.AccumulatedRewards)

	// User1 Claimable: (8.5 - 8.5) * 60M = 0
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "User1 claimable should be 0. Got %s", result.Rewards)

	// New epoch: +100 DYM (This is Epoch 8 for distributions)
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify Endorsement:
	// Prev GA = 8.5. RewardsPerShare = 100M / 60M = 1.666...
	// New GA = 8.5 + 1.666... = 10.1666...
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)

	rewardsPerShareVal := math.LegacyNewDecFromInt(commontypes.DYM.MulRaw(100)).QuoTruncate(expectedUser1SharesDec)
	expectedAccumulatorRepeating := expectedAccumulatorEpoch7.Add(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, rewardsPerShareVal))
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulatorRepeating), "GA should be ~10.166. Expected %s, got %s", expectedAccumulatorRepeating, endorsement.Accumulator)
	s.Require().True(endorsement.TotalShares.Equal(expectedUser1SharesDec), "TotalShares should still be 60M. Expected %s, got %s", expectedUser1SharesDec, endorsement.TotalShares)
	unlockedCoins = endorsement.TotalCoins.Sub(endorsement.DistributedCoins...) // Now 0 + 100 DYM
	s.Require().True(unlockedCoins.Equal(dym100), "Unlocked coins should be 100 DYM. Expected %s, got %s", dym100, unlockedCoins)

	// Verify User1 Position (LSA and AUR are not changed by epoch start itself)
	posUser1, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user1Addr, rollappID)
	s.Require().NoError(err)
	s.Require().True(posUser1.Shares.Equal(expectedUser1SharesDec), "User1 shares should remain 60M.")
	s.Require().True(posUser1.LastSeenAccumulator.Equal(expectedAccumulatorEpoch7), "User1 LSA should remain 8.5 (updates on action).")
	s.Require().True(posUser1.AccumulatedRewards.IsZero(), "User1 AUR should remain 0.")

	// User1 Claimable: (NewGA - LSA) * Shares = (10.166... - 8.5) * 60M = 1.666... * 60M = 99M (due to truncation)
	expectedClaimableUser1Repeating := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(99)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, rollappGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableUser1Repeating), "User1 claimable should be 99 DYM. Expected %s, got %s", expectedClaimableUser1Repeating, result.Rewards)
}

func (s *KeeperTestSuite) BeginEpoch(epochID string) {
	info := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochID)
	s.Ctx = s.Ctx.WithBlockTime(info.CurrentEpochStartTime.Add(info.Duration).Add(time.Minute))
	s.App.EpochsKeeper.BeginBlocker(s.Ctx)
}
