package keeper_test

import (
	"errors"
	"time"

	"cosmossdk.io/collections"
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
	initial1 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(40))
	del1 := s.CreateDelegator(valAddr, initial1)

	// User2 delegates 100 DYM (but only endorses with 60 shares in step 3)
	initial2 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))
	del2 := s.CreateDelegator(valAddr, initial2)

	// Helper to create 100 DYM coins
	dym100 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	dym200 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(200)))

	/***************************************************************/
	/*                           Epoch 1                           */
	/***************************************************************/

	// User1 endorses with 40 shares
	// User1 votes 100% on the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del1.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// Verify initial state: accumulator = 0.0, total shares = 40
	endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.IsZero(), "accumulator should be 0.0")
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(40))), "total shares should be 40")

	// User1 should have 0 claimable rewards (nothing distributed yet)
	user1Addr := sdk.MustAccAddressFromBech32(del1.GetDelegatorAddr())
	result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
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

	// User1 should have 100 DYM claimable: (2.5 - 0) * 40 = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
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

	// User2 should have 0 claimable (LSA = 2.5, so (2.5 - 2.5) * 60 = 0)
	user2Addr := sdk.MustAccAddressFromBech32(del2.GetDelegatorAddr())
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user2 should have 0 claimable rewards initially")

	// User1 should still have 100 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
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

	// User1 should have 140 DYM claimable: (3.5 - 0) * 40 = 140
	dym140 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(140)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym140), "user1 should have 140 DYM claimable")

	// User2 should have 60 DYM claimable: (3.5 - 2.5) * 60 = 60
	dym60 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(60)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should have 60 DYM claimable")

	// Claim User1 and verify their balance
	{
		// Snapshot User1 balance before claiming
		beforeBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 claims
		err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, endorsementGaugeID)
		s.Require().NoError(err)

		// Snapshot User1 balance after claiming
		afterBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 should have increased by 140 DYM
		s.Require().True(afterBalance.Sub(beforeBalance).Amount.Equal(dym140[0].Amount), "user1 balance should have increased by 140 DYM")
	}

	// User1 should now have 0 claimable (LSA updated to 3.5)
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable after claiming")

	// User2 should still have 60 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should still have 60 DYM claimable")

	// User2 un-endorses
	s.RevokeVote(types.MsgRevokeVote{
		Voter: user2Addr.String(),
	})

	// User2 should still have 60 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should still have 60 DYM claimable")

	// Verify total shares = 40 (only user1 left)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(40))), "total shares should be 40 after user2 leaves")

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

	// User1 should have 100 DYM claimable: (6 - 3.5) * 40 = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID)
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
		err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, endorsementGaugeID)
		s.Require().NoError(err)

		// Snapshot User1 balance after claiming
		afterBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

		// User1 should have increased by 100 DYM
		s.Require().True(afterBalance.Sub(beforeBalance).Amount.Equal(dym100[0].Amount), "user1 balance should have increased by 100 DYM")
	}

	// Verify total shares = 0 (no endorsers left)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.IsZero(), "total shares should be 0 after all users leave")

	// User2 re-endorses with 100 shares
	// User2 now votes with 100 DYM equivalent
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// User2 should have 60 DYM claimable bc of accumulated rewards
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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

	// User2 should have 160 DYM claimable: (new_accumulator - 6.0) * shares + accumulated = 160
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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

	// Verify AUR becomes 160 DYM, LSA becomes 7.0, shares become 300
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	expectedAUR := dym100.Add(dym60...) // 160 DYM
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator)
	s.Require().Truef(pos.Shares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(300))), "Shares mismatch. Expected %s, got %s", math.LegacyNewDecFromInt(types.DYM.MulRaw(300)), pos.Shares)

	// User2 should have 160 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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

	// Verify User2 Position: AUR=160, LSA=7.0, Shares=200
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator)
	s.Require().Truef(pos.Shares.Equal(dym200SharesDec), "Shares mismatch. Expected %s, got %s", dym200SharesDec, pos.Shares)

	// User2 should still have 160 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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
	expectedAccumulator_Epoch6 := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(75, 1))) // 7.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator_Epoch6), "GA should be 7.5 in Epoch 6")
	s.Require().True(endorsement.TotalShares.Equal(dym200SharesDec), "TotalShares should be 200 in Epoch 6")

	// Verify User2 Position (unchanged by epoch start): AUR=160, LSA=7.0, Shares=200
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator, pos.LastSeenAccumulator) // LSA is still 7.0
	s.Require().Truef(pos.Shares.Equal(dym200SharesDec), "Shares mismatch. Expected %s, got %s", dym200SharesDec, pos.Shares)

	// User2 Claimable: (7.5 - 7.0) * 200 + 160 = 100 + 160 = 260 DYM
	expectedClaimableEpoch6 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(260)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch6), "User2 claimable should be 260 DYM in Epoch 6")

	/***************************************************************/
	/*               User2 removes 100 shares (updates vote)       */
	/***************************************************************/
	// User2 (total stake 200) updates vote from 200 shares to 100 shares
	dym100SharesWeight := commontypes.DYM.MulRaw(100)
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(), // user2Addr.String()
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: dym100SharesWeight}, // Vote with 100 shares for rollapp gauge 1
		},
	})

	// Verify Endorsement: GA=7.5, TotalShares = 100 (200(old total) - 200(old U2) + 100(new U2))
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator_Epoch6), "GA should be 7.5 after vote update")
	dym100SharesDec := math.LegacyNewDecFromInt(types.DYM.MulRaw(100))
	s.Require().True(endorsement.TotalShares.Equal(dym100SharesDec), "TotalShares should be 100 after vote update")

	// Verify User2 Position:
	// Rewards to bank: (7.5 - 7.0) * 200 = 100. New AUR = 160 + 100 = 260. LSA = 7.5. Shares = 100.
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch6.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch6, pos.AccumulatedRewards) // AUR is 260
	s.Require().Truef(expectedAccumulator_Epoch6.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator_Epoch6, pos.LastSeenAccumulator) // LSA is 7.5
	s.Require().Truef(pos.Shares.Equal(dym100SharesDec), "Shares mismatch. Expected %s, got %s", dym100SharesDec, pos.Shares)

	// User2 Claimable: (7.5 - 7.5) * 100 + 260 = 260 DYM
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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
	expectedAccumulator_Epoch7 := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(85, 1))) // 8.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator_Epoch7), "GA should be 8.5 in Epoch 7")
	s.Require().True(endorsement.TotalShares.Equal(dym100SharesDec), "TotalShares should be 100 in Epoch 7")

	// Verify User2 Position (AUR and LSA unchanged by epoch start): AUR=260, LSA=7.5, Shares=100
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch6.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch6, pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator_Epoch6.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator_Epoch6, pos.LastSeenAccumulator) // LSA is still 7.5
	s.Require().Truef(pos.Shares.Equal(dym100SharesDec), "Shares mismatch. Expected %s, got %s", dym100SharesDec, pos.Shares)

	// User2 Claimable: (8.5 - 7.5) * 100 + 260 = 100 + 260 = 360 DYM
	expectedClaimableEpoch7 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(360)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
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
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator_Epoch7), "GA should be 8.5 after unstake all")
	s.Require().True(endorsement.TotalShares.IsZero(), "TotalShares should be 0 after unstake all")

	// Verify User2 Position:
	// Rewards to bank: (8.5 - 7.5) * 100 = 100. New AUR = 260 + 100 = 360. LSA = 8.5. Shares = 0.
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().Truef(expectedClaimableEpoch7.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedClaimableEpoch7, pos.AccumulatedRewards) // AUR is 360
	s.Require().Truef(expectedAccumulator_Epoch7.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator_Epoch7, pos.LastSeenAccumulator) // LSA is 8.5
	s.Require().True(pos.Shares.IsZero(), "Shares mismatch. Expected 0, got %s", pos.Shares)

	// User2 Claimable: (8.5 - 8.5) * 0 + 360 = 360 DYM
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedClaimableEpoch7), "User2 claimable should be 360 DYM after unstake all")

	/***************************************************************/
	/*                         User2 claims                        */
	/***************************************************************/
	balanceBeforeClaim := s.App.BankKeeper.GetBalance(s.Ctx, user2Addr, sdk.DefaultBondDenom)
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	balanceAfterClaim := s.App.BankKeeper.GetBalance(s.Ctx, user2Addr, sdk.DefaultBondDenom)
	s.Require().True(balanceAfterClaim.Sub(balanceBeforeClaim).Amount.Equal(expectedClaimableEpoch7[0].Amount), "User2 claimed amount mismatch")

	// Verify User2 Position: AUR=0, LSA=8.5, Shares=0
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, user2Addr, rollappID)
	s.Require().NoError(err)
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero after claim. Got %s", pos.AccumulatedRewards)
	s.Require().Truef(expectedAccumulator_Epoch7.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", expectedAccumulator_Epoch7, pos.LastSeenAccumulator)
	s.Require().True(pos.Shares.IsZero(), "Shares should be zero after claim. Got %s", pos.Shares)

	// User2 Claimable should be 0
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "User2 claimable should be zero after claim")
}

// TestEndorsementsMultiCurrency tests the lazy accumulator with multiple currencies
func (s *KeeperTestSuite) TestEndorsementsMultiCurrency() {
	// Create a rollapp gauge (ID 1)
	rollappID := s.CreateDefaultRollapp()

	// Create validators and delegators
	val := s.CreateValidator()
	valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
	s.Require().NoError(err)

	// User1 delegates 50 DYM
	initial1 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(50))
	del1 := s.CreateDelegator(valAddr, initial1)

	// User2 delegates 50 DYM
	initial2 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(50))
	del2 := s.CreateDelegator(valAddr, initial2)

	// Both users endorse the rollapp
	s.Vote(types.MsgVote{
		Voter: del1.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// Add multi-currency rewards: 100 DYM + 50 USDC
	multiCurrencyRewards := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)), // 100 DYM
		sdk.NewCoin("usdc", commontypes.DYM.MulRaw(50)),                // 50 USDC
	)

	err = s.App.SponsorshipKeeper.UpdateEndorsementTotalCoins(s.Ctx, rollappID, multiCurrencyRewards)
	s.Require().NoError(err)

	// Verify accumulator has both currencies
	endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)

	// Expected accumulator: DYM = 100/100 = 1.0, USDC = 50/100 = 0.5
	expectedAccumulator := sdk.NewDecCoins(
		sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(1)), // 1.0 DYM
		sdk.NewDecCoinFromDec("usdc", math.LegacyNewDecWithPrec(5, 1)),    // 0.5 USDC
	)
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should have both currencies")

	// Each user should have 50 DYM + 25 USDC claimable: (1.0 - 0) * 50 = 50 DYM, (0.5 - 0) * 50 = 25 USDC
	expectedRewards := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(50)), // 50 DYM
		sdk.NewCoin("usdc", commontypes.DYM.MulRaw(25)),               // 25 USDC
	)

	user1Addr := sdk.MustAccAddressFromBech32(del1.GetDelegatorAddr())
	result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedRewards), "user1 should have multi-currency rewards")

	user2Addr := sdk.MustAccAddressFromBech32(del2.GetDelegatorAddr())
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedRewards), "user2 should have multi-currency rewards")

	// User1 claims
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)

	// User1 should now have 0 claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, endorsementGaugeID) // Use correct gauge ID
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable after claiming")

	// User2 should still have the same claimable amount
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, endorsementGaugeID) // Use correct gauge ID
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedRewards), "user2 should still have multi-currency rewards")
}

func (s *KeeperTestSuite) BeginEpoch(epochID string) {
	info := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochID)
	s.Ctx = s.Ctx.WithBlockTime(info.CurrentEpochStartTime.Add(info.Duration).Add(time.Minute))
	s.App.EpochsKeeper.BeginBlocker(s.Ctx)
}

// createEndorsementGauge is a helper function to create an endorsement gauge.
func (s *KeeperTestSuite) createEndorsementGauge(rollappID string, creator sdk.AccAddress, rewards sdk.Coins, numEpochsPaidOver uint64) uint64 {
	gaugeID, err := s.App.IncentivesKeeper.CreateEndorsementGauge(
		s.Ctx,
		false, // perpetual = false
		creator,
		rewards,
		incentivestypes.EndorsementGauge{
			RollappId: rollappID,
		},
		s.Ctx.BlockTime(),
		numEpochsPaidOver,
	)
	s.Require().NoError(err)
	return gaugeID
}

// setupEndorsementScenario directly sets up the initial state for an endorsement and an endorser's position.
func (s *KeeperTestSuite) setupEndorsementScenario(
	rollappID string,
	userAddr sdk.AccAddress,
	userShares math.LegacyDec,
	userLSA sdk.DecCoins,
	userAUR sdk.Coins,
	endorsementAccumulator sdk.DecCoins,
	endorsementTotalShares math.LegacyDec,
	endorsementGaugeID uint64,
) {
	k := s.App.SponsorshipKeeper
	ctx := s.Ctx

	// Save Endorsement
	endorsement := types.Endorsement{
		RollappId:        rollappID,
		RollappGaugeId:   endorsementGaugeID,
		Accumulator:      endorsementAccumulator,
		TotalShares:      endorsementTotalShares,
		TotalCoins:       sdk.NewCoins(), // Not critical for LSA/AUR logic testing here
		DistributedCoins: sdk.NewCoins(), // Not critical for LSA/AUR logic testing here
	}
	err := k.SaveEndorsement(ctx, endorsement)
	s.Require().NoError(err)

	// Save EndorserPosition
	// If userShares is zero, it implies the user might not have a position yet, or has zero shares.
	// If a position is to be created, it starts with these explicit values.
	endorserPosition := types.EndorserPosition{
		Shares:              userShares,
		LastSeenAccumulator: userLSA,
		AccumulatedRewards:  userAUR,
	}
	err = k.SaveEndorserPosition(ctx, userAddr, rollappID, endorserPosition)
	s.Require().NoError(err)
}

// updateEndorserShares simulates a user updating their shares for an endorsement.
// It correctly updates AUR, LSA, and total endorsement shares.
// This function assumes the user might be new or existing.
func (s *KeeperTestSuite) updateEndorserShares(
	rollappID string,
	userAddr sdk.AccAddress,
	newTotalUserShares math.LegacyDec,
	_ uint64, // endorsementGaugeID, kept for signature consistency if needed later
) {
	k := s.App.SponsorshipKeeper
	ctx := s.Ctx

	endorsement, err := k.GetEndorsement(ctx, rollappID)
	s.Require().NoError(err)

	endorserPosition, err := k.GetEndorserPosition(ctx, userAddr, rollappID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// This is a new endorser for this rollapp
			endorserPosition = types.NewDefaultEndorserPosition()
		} else {
			s.Require().NoError(err, "Failed to get endorser position")
			return // Should not happen due to Require.NoError
		}
	}

	// Calculate rewards to bank with old shares before updating them
	// For a brand new position, Shares and LSA are zero, so rewardsToBank will be zero.
	rewardsToBank := endorserPosition.RewardsToBank(endorsement.Accumulator)
	endorserPosition.AccumulatedRewards = endorserPosition.AccumulatedRewards.Add(rewardsToBank...)

	// Update LSA to current global accumulator
	endorserPosition.LastSeenAccumulator = endorsement.Accumulator

	// Update total shares in endorsement: subtract old shares, add new shares
	endorsement.TotalShares = endorsement.TotalShares.Sub(endorserPosition.Shares).Add(newTotalUserShares)
	if endorsement.TotalShares.IsNegative() {
		s.T().Fatalf("Endorsement total shares became negative: %s", endorsement.TotalShares)
	}

	// Update user's shares to the new total
	endorserPosition.Shares = newTotalUserShares

	err = k.SaveEndorserPosition(ctx, userAddr, rollappID, endorserPosition)
	s.Require().NoError(err)
	err = k.SaveEndorsement(ctx, endorsement)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestScenario_StakeUnstakeAndClaim() {
	// Initial Setup
	rollappID := s.CreateDefaultRollapp()
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10000))))
	endorsementGaugeID := s.createEndorsementGauge(rollappID, gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000))), 10)

	userAddr := apptesting.CreateRandomAccounts(1)[0]

	// Initial state from plan: User has 100 shares, LSA = 6.0, AUR = 0. GA = 7.0. TotalShares = 100.
	initialUserShares := math.LegacyNewDec(100)
	initialLSA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(6)))
	initialAUR := sdk.NewCoins()
	initialGA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(7)))
	initialTotalEndorsementShares := math.LegacyNewDec(100)

	s.setupEndorsementScenario(rollappID, userAddr, initialUserShares, initialLSA, initialAUR, initialGA, initialTotalEndorsementShares, endorsementGaugeID)

	// --- Action 1: User stakes 100 more (total shares become 200) ---
	s.updateEndorserShares(rollappID, userAddr, math.LegacyNewDec(200), endorsementGaugeID)

	// Verify AUR becomes 100 DYM, LSA becomes 7.0
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	expectedAUR := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	s.Require().True(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().True(initialGA.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", initialGA, pos.LastSeenAccumulator)

	// Verify claimable (EstimateClaim = (GA-LSA)*Shares + AUR)
	// (7.0 - 7.0) * 200 + 100 = 100
	claimable, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(expectedAUR.Equal(claimable.Rewards), "Claimable mismatch. Expected %s, got %s", expectedAUR, claimable.Rewards)

	// --- Action 2: User unstakes 50 (total shares become 150) ---
	s.updateEndorserShares(rollappID, userAddr, math.LegacyNewDec(150), endorsementGaugeID)

	// Verify AUR remains 100 DYM, LSA remains 7.0
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch after unstake. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)
	s.Require().True(initialGA.Equal(pos.LastSeenAccumulator), "LSA mismatch after unstake. Expected %s, got %s", initialGA, pos.LastSeenAccumulator)

	// Verify claimable remains 100
	// (7.0 - 7.0) * 150 + 100 = 100
	claimable, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(expectedAUR.Equal(claimable.Rewards), "Claimable mismatch after unstake. Expected %s, got %s", expectedAUR, claimable.Rewards)

	// --- Action 3: New epoch: +100 DYM ---
	err = s.App.SponsorshipKeeper.UpdateEndorsementTotalCoins(s.Ctx, rollappID, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))))
	s.Require().NoError(err)

	// Verify new GA: 7.0 + (100 / 150) = 7.666...
	endorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedNewGAVal := math.LegacyNewDec(7).Add(math.LegacyNewDec(100).Quo(math.LegacyNewDec(150)))
	expectedNewGA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, expectedNewGAVal))
	s.Require().True(expectedNewGA.Equal(endorsement.Accumulator), "New GA mismatch. Expected %s, got %s", expectedNewGA, endorsement.Accumulator)

	// Verify claimable becomes 200 DYM: ((7.666... - 7.0) * 150) + 100 (AUR) = 100 + 100 = 200
	// AUR should still be 100
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(expectedAUR.Equal(pos.AccumulatedRewards), "AUR should remain 100 before claim. Got %s", pos.AccumulatedRewards)

	expectedClaimableAfterEpoch := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(200)))
	claimable, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(expectedClaimableAfterEpoch.Equal(claimable.Rewards), "Claimable after epoch mismatch. Expected %s, got %s", expectedClaimableAfterEpoch, claimable.Rewards)

	// --- Action 4: User claims ---
	balanceBeforeClaim := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, sdk.DefaultBondDenom)
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	balanceAfterClaim := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, sdk.DefaultBondDenom)
	s.Require().True(balanceAfterClaim.Sub(balanceBeforeClaim).Amount.Equal(commontypes.DYM.MulRaw(200)), "Claimed amount mismatch")

	// Verify AUR becomes 0, LSA updates to current GA
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero after claim. Got %s", pos.AccumulatedRewards)
	s.Require().True(expectedNewGA.Equal(pos.LastSeenAccumulator), "LSA should update to new GA. Expected %s, got %s", expectedNewGA, pos.LastSeenAccumulator)

	// Verify claimable becomes 0
	claimable, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(claimable.Rewards.IsZero(), "Claimable should be zero after claim. Got %s", claimable.Rewards)
}

func (s *KeeperTestSuite) TestScenario_UnstakeAll() {
	rollappID := s.CreateDefaultRollapp()
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10000))))
	endorsementGaugeID := s.createEndorsementGauge(rollappID, gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000))), 10)
	userAddr := apptesting.CreateRandomAccounts(1)[0]

	// Initial state: User has 100 shares, LSA = 6.0, AUR = 0. GA = 7.0. TotalShares = 100.
	s.setupEndorsementScenario(
		rollappID, userAddr,
		math.LegacyNewDec(100), // userShares
		sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(6))), // userLSA
		sdk.NewCoins(), // userAUR
		sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(7))), // endorsementAccumulator (GA)
		math.LegacyNewDec(100), // endorsementTotalShares
		endorsementGaugeID,
	)

	currentGA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(7)))

	// --- Action 1: User unstakes all 100 shares ---
	s.updateEndorserShares(rollappID, userAddr, math.LegacyZeroDec(), endorsementGaugeID) // New shares = 0

	// Verify AUR becomes 100 DYM ((7-6)*100)
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	expectedAUR := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	s.Require().True(expectedAUR.Equal(pos.AccumulatedRewards), "AUR mismatch. Expected %s, got %s", expectedAUR, pos.AccumulatedRewards)

	// Verify user shares become 0
	s.Require().True(pos.Shares.IsZero(), "User shares should be zero. Got %s", pos.Shares)

	// Verify LSA becomes 7.0
	s.Require().True(currentGA.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", currentGA, pos.LastSeenAccumulator)

	// Verify claimable (EstimateClaim = (GA-LSA)*Shares + AUR)
	// (7.0 - 7.0) * 0 + 100 = 100
	claimable, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(expectedAUR.Equal(claimable.Rewards), "Claimable mismatch. Expected %s, got %s", expectedAUR, claimable.Rewards)

	// --- Action 2: User claims ---
	balanceBeforeClaim := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, sdk.DefaultBondDenom)
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	balanceAfterClaim := s.App.BankKeeper.GetBalance(s.Ctx, userAddr, sdk.DefaultBondDenom)
	s.Require().True(balanceAfterClaim.Sub(balanceBeforeClaim).Amount.Equal(commontypes.DYM.MulRaw(100)), "Claimed amount mismatch")

	// Verify AUR becomes 0
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero after claim. Got %s", pos.AccumulatedRewards)
}

func (s *KeeperTestSuite) TestScenario_EndorseWithNoPriorVote() {
	rollappID := s.CreateDefaultRollapp()
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10000))))
	endorsementGaugeID := s.createEndorsementGauge(rollappID, gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000))), 10)
	userAddr := apptesting.CreateRandomAccounts(1)[0]

	// Initial state: GA = 7.0. Endorsement has 100 total shares (from other users, not this one).
	// User has no prior vote/shares, so no EndorserPosition object exists for them yet.
	initialGA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(7)))
	initialTotalEndorsementShares := math.LegacyNewDec(100) // Shares from other users

	// Setup only the endorsement state, user has no position yet.
	endorsement := types.Endorsement{
		RollappId:      rollappID,
		RollappGaugeId: endorsementGaugeID,
		Accumulator:    initialGA,
		TotalShares:    initialTotalEndorsementShares,
	}
	err := s.App.SponsorshipKeeper.SaveEndorsement(s.Ctx, endorsement)
	s.Require().NoError(err)

	// --- Action 1: User endorses with 100 shares ---
	// This user is new, so their old shares for this endorsement were 0.
	s.updateEndorserShares(rollappID, userAddr, math.LegacyNewDec(100), endorsementGaugeID)

	// Verify user shares become 100
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(math.LegacyNewDec(100).Equal(pos.Shares), "User shares mismatch. Expected 100, got %s", pos.Shares)

	// Verify LSA becomes 7.0 (current GA)
	s.Require().True(initialGA.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", initialGA, pos.LastSeenAccumulator)

	// Verify AUR is 0
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero for new endorser. Got %s", pos.AccumulatedRewards)

	// Verify claimable is 0 ( (7.0-7.0)*100 + 0 = 0 )
	claimable, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(claimable.Rewards.IsZero(), "Claimable should be zero. Got %s", claimable.Rewards)

	// Verify total endorsement shares increased by 100
	updatedEndorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedTotalShares := initialTotalEndorsementShares.Add(math.LegacyNewDec(100))
	s.Require().True(expectedTotalShares.Equal(updatedEndorsement.TotalShares), "Total endorsement shares mismatch. Expected %s, got %s", expectedTotalShares, updatedEndorsement.TotalShares)
}

func (s *KeeperTestSuite) TestScenario_EndorseEmptyEndorsement() {
	rollappID := s.CreateDefaultRollapp()
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(10000))))
	endorsementGaugeID := s.createEndorsementGauge(rollappID, gaugeCreator, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000))), 10)
	userAddr := apptesting.CreateRandomAccounts(1)[0]

	// Initial state: First epoch. GA = 0. Total shares for the endorsement = 0.
	initialGA := sdk.NewDecCoins() // Empty or zero
	initialTotalEndorsementShares := math.LegacyZeroDec()

	// Setup only the endorsement state, user has no position yet.
	// Ensure GA is explicitly zero for the bond denom if DecCoins() is not specific enough
	initialGA = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyZeroDec()))

	endorsementToSave := types.Endorsement{
		RollappId:      rollappID,
		RollappGaugeId: endorsementGaugeID,
		Accumulator:    initialGA,
		TotalShares:    initialTotalEndorsementShares,
	}
	err := s.App.SponsorshipKeeper.SaveEndorsement(s.Ctx, endorsementToSave)
	s.Require().NoError(err)

	// --- Action 1: User endorses with 100 shares ---
	s.updateEndorserShares(rollappID, userAddr, math.LegacyNewDec(100), endorsementGaugeID)

	// Verify user shares become 100
	pos, err := s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(math.LegacyNewDec(100).Equal(pos.Shares), "User shares mismatch. Expected 100, got %s", pos.Shares)

	// Verify LSA becomes 0 (current GA)
	s.Require().True(initialGA.Equal(pos.LastSeenAccumulator), "LSA mismatch. Expected %s, got %s", initialGA, pos.LastSeenAccumulator)

	// Verify AUR is 0
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should be zero. Got %s", pos.AccumulatedRewards)

	// Verify claimable is 0
	claimable, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(claimable.Rewards.IsZero(), "Claimable should be zero. Got %s", claimable.Rewards)

	// Verify total shares for the endorsement become 100
	currentEndorsement, err := s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(math.LegacyNewDec(100).Equal(currentEndorsement.TotalShares), "Total endorsement shares mismatch. Expected 100, got %s", currentEndorsement.TotalShares)

	// Verify GA remains 0
	s.Require().True(initialGA.Equal(currentEndorsement.Accumulator), "GA should remain zero. Expected %s, got %s", initialGA, currentEndorsement.Accumulator)

	// --- Action 2: New epoch: +100 DYM ---
	err = s.App.SponsorshipKeeper.UpdateEndorsementTotalCoins(s.Ctx, rollappID, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))))
	s.Require().NoError(err)

	// Verify new GA: 0 + (100 / 100) = 1.0
	currentEndorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedNewGA := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDec(1)))
	s.Require().True(expectedNewGA.Equal(currentEndorsement.Accumulator), "New GA mismatch. Expected %s, got %s", expectedNewGA, currentEndorsement.Accumulator)

	// Verify claimable becomes 100 DYM ((1.0 - 0) * 100) + 0 (AUR) = 100
	expectedClaimable := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))
	claimable, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, userAddr, endorsementGaugeID)
	s.Require().NoError(err)
	s.Require().True(expectedClaimable.Equal(claimable.Rewards), "Claimable after epoch mismatch. Expected %s, got %s", expectedClaimable, claimable.Rewards)

	// Verify AUR remains 0 (as it wasn't touched by UpdateEndorsementTotalCoins, and claim hasn't happened)
	pos, err = s.App.SponsorshipKeeper.GetEndorserPosition(s.Ctx, userAddr, rollappID)
	s.Require().NoError(err)
	s.Require().True(pos.AccumulatedRewards.IsZero(), "AUR should still be zero. Got %s", pos.AccumulatedRewards)
}
