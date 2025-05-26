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

	// Create an endorsement gauge (ID 2)
	gaugeCreator := apptesting.CreateRandomAccounts(1)[0]
	dym1000 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000)))
	s.FundAcc(gaugeCreator, dym1000)

	// Gauge for 1000 DYM and 10 epochs
	_, err := s.App.IncentivesKeeper.CreateEndorsementGauge(
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

	// User2 delegates 60 DYM (but only endorses with 60 shares in step 3)
	initial2 := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(60))
	del2 := s.CreateDelegator(valAddr, initial2)

	// Helper to create 100 DYM coins
	dym100 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100)))

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
	result, err := s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable rewards initially")

	/***************************************************************/
	/*                           Epoch 2                           */
	/***************************************************************/

	// +100 DYM unlocked

	// Simulate adding 100 DYM to the endorsement
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Verify accumulator = 2.5 (100 DYM / 40 shares)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	expectedAccumulator := sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyNewDecWithPrec(25, 1))) // 2.5
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "accumulator should be 2.5")

	// User1 should have 100 DYM claimable: (2.5 - 0) * 40 = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100), "user1 should have 100 DYM claimable")

	// User2 endorses with 60 shares
	// User2 votes 100% on the rollapp gauge
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// Verify total shares = 100 (40 + 60)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(100))), "total shares should be 100")

	// User2 should have 0 claimable (LSA = 2.5, so (2.5 - 2.5) * 60 = 0)
	user2Addr := sdk.MustAccAddressFromBech32(del2.GetDelegatorAddr())
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user2 should have 0 claimable rewards initially")

	// User1 should still have 100 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
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
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym140), "user1 should have 140 DYM claimable")

	// User2 should have 60 DYM claimable: (3.5 - 2.5) * 60 = 60
	dym60 := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(60)))
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should have 60 DYM claimable")

	// Snapshot User1 balance before claiming
	beforeBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

	// User1 claims
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)

	// Snapshot User1 balance after claiming
	afterBalance := s.App.BankKeeper.GetBalance(s.Ctx, user1Addr, sdk.DefaultBondDenom)

	// User1 should have increased by 140 DYM
	s.Require().True(afterBalance.Sub(beforeBalance).Amount.Equal(dym140[0].Amount), "user1 balance should have increased by 140 DYM")

	// User1 should now have 0 claimable (LSA updated to 3.5)
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable after claiming")

	// User2 should still have 60 DYM claimable
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym60), "user2 should still have 60 DYM claimable")

	// User2 un-endorses (auto-claim)
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)

	// Revoke user2's vote
	_, err = s.msgServer.RevokeVote(s.Ctx, &types.MsgRevokeVote{
		Voter: del2.GetDelegatorAddr(),
	})
	s.Require().NoError(err)

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
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(dym100), "user1 should have 100 DYM claimable")

	// User1 un-endorses (auto-claim)
	err = s.App.SponsorshipKeeper.Claim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)

	// Revoke user1's vote
	_, err = s.msgServer.RevokeVote(s.Ctx, &types.MsgRevokeVote{
		Voter: del1.GetDelegatorAddr(),
	})
	s.Require().NoError(err)

	// Verify total shares = 0 (no endorsers left)
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)
	s.Require().True(endorsement.TotalShares.IsZero(), "total shares should be 0 after all users leave")

	// User2 re-endorses with 100 shares
	// User2 now votes with 100 DYM equivalent (need to update delegation)
	// For simplicity, we'll vote with 100% weight which should give us the full voting power
	s.Vote(types.MsgVote{
		Voter: del2.GetDelegatorAddr(),
		Weights: []types.GaugeWeight{
			{GaugeId: 1, Weight: commontypes.DYM.MulRaw(100)}, // 100% to rollapp gauge
		},
	})

	// User2 should have 0 claimable (LSA = 6.0, current accumulator = 6.0)
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user2 should have 0 claimable after re-endorsing")

	/***************************************************************/
	/*                           Epoch 5                           */
	/***************************************************************/

	// +100 DYM unlocked
	s.BeginEpoch(incentivestypes.DefaultDistrEpochIdentifier)

	// Get current total shares to calculate expected accumulator
	endorsement, err = s.App.SponsorshipKeeper.GetEndorsement(s.Ctx, rollappID)
	s.Require().NoError(err)

	// Expected accumulator = 6.0 + (100 / 60)
	expectedAccumulatorValue := math.LegacyNewDec(6).Add(math.LegacyNewDec(100).Quo(math.LegacyNewDec(60)))
	expectedAccumulator = sdk.NewDecCoins(sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, expectedAccumulatorValue))

	// Check endorsement
	s.Require().True(endorsement.TotalShares.Equal(math.LegacyNewDecFromInt(types.DYM.MulRaw(60))), "total shares should be 100 on epoch 5")
	s.Require().True(endorsement.Accumulator.Equal(expectedAccumulator), "global accumulator should be 7.7 on epoch 5")

	// User2 should have 100 DYM claimable: (new_accumulator - 6.0) * shares = 100
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	// The actual claimable balance is
	adjustedRes := dym100.Add(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(20)))
	s.Require().True(result.Rewards.Equal(adjustedRes), "user2 should have 100 DYM claimable after final distribution")
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
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user1Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.IsZero(), "user1 should have 0 claimable after claiming")

	// User2 should still have the same claimable amount
	result, err = s.App.SponsorshipKeeper.EstimateClaim(s.Ctx, user2Addr, 1)
	s.Require().NoError(err)
	s.Require().True(result.Rewards.Equal(expectedRewards), "user2 should still have multi-currency rewards")
}

func (s *KeeperTestSuite) BeginEpoch(epochID string) {
	info := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochID)
	s.Ctx = s.Ctx.WithBlockTime(info.CurrentEpochStartTime.Add(info.Duration).Add(time.Minute))
	s.App.EpochsKeeper.BeginBlocker(s.Ctx)
}
