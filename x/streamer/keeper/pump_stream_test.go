package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Test case parameters structure
type pumpTestCase struct {
	name              string
	pumpParams        *types.MsgCreateStream_PumpParams
	numEpochsPaidOver uint64
	epochIdentifier   string
	streamCoins       sdk.Coins
}

func (s *KeeperTestSuite) TestShouldPump() {
	b, err := s.App.StreamerKeeper.EpochBlocks(s.Ctx, "day")
	s.Require().NoError(err)

	pumpNum := uint64(5000)

	s.Run("GenerateUnifiedRandom", func() {
		// Pump hash
		ctx := hashPump(s.Ctx)
		r1 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandom(ctx, b.BigIntMut()),
		) //  510461966151279

		// No pump hash
		ctx = hashNoPump(s.Ctx)
		r2 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandom(ctx, b.BigIntMut()),
		) //  723264247167302

		middle := math.NewIntFromUint64(pumpNum)

		s.Require().True(r1.LT(middle), "expected r1 < middle, got: %s < %s", r1, middle)
		s.Require().True(middle.LT(r2), "expected middle < r2, got: %s < %s ", middle, r2)
	})

	s.Run("ShouldPump", func() {
		// Pump hash should pump
		ctx := hashPump(s.Ctx)
		pumpAmt, err := keeper.ShouldPump(
			ctx,
			commontypes.DYM.MulRaw(10),
			commontypes.DYM.MulRaw(10),
			pumpNum,
			b,
		)
		s.Require().NoError(err)
		s.Require().False(pumpAmt.IsZero())
		expectedAmt := math.NewInt(2040061966151279)
		s.Require().True(expectedAmt.Equal(pumpAmt), "expected %s, got: %s", expectedAmt, pumpAmt)

		// No pump hash should not pump
		ctx = hashNoPump(s.Ctx)
		pumpAmt, err = keeper.ShouldPump(
			ctx,
			commontypes.DYM.MulRaw(10),
			commontypes.DYM.MulRaw(10),
			pumpNum,
			b,
		)
		s.Require().NoError(err)
		s.Require().True(pumpAmt.IsZero())
	})
}

func (s *KeeperTestSuite) TestPumpStream() {
	// Define test cases with different parameters
	testCases := []pumpTestCase{
		{
			name: "pump stream scenario",
			pumpParams: &types.MsgCreateStream_PumpParams{
				NumTopRollapps: 2,
				NumPumps:       5000, // This is a hardcoded number for test convenience
			},
			numEpochsPaidOver: 10,
			epochIdentifier:   "day",
			streamCoins:       sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.runPumpStreamTest(tc)
		})
	}
}

// Scenario:
//  1. Create 4 rollapps
//  2. Create 2 users (delegators) who have staking power. Both delegate 100 DYM.
//  3. Vote on rollapps. User 1 votes 60 on RA1 and 40 on RA2. User 2 votes 60 on RA2 and 40 on RA3. RA4 is not endorsed.
//  4. Create an IRO for RA1 and RA2. LiquidityDenom is BaseDenom.
//  5. Create a Pump Stream for 10 epochs for 100 DYM total with 2 RAs to pump => 10 DYM on every epoch.
//  6. Validate the Pump Stream – It is not active and must have only DYM in the coin list, EpochBudget and EpochBudgetLeft are 0.
//  7. Simulate an epoch start.
//  8. Validate the Pump Stream – EpochBudget and EpochBudgetLeft are changed = 10 DYM.
//  9. Use a predictable block hash in sdk.Context that the pump won’t definitely happen.
//  10. Simulate a new block and verify that the pump wasn’t executed:
//     – DistributedCoins, EpochBudgetLeft, IRO.SoldAmt are the same
//     - No EventPumped event
//     - x/streamer balance is the same
//  11. Put a predictable block hash in sdk.Context that the pump will definitely happen.
//  12. Simulate a new block and verify that the pump was executed:
//     - TopRollappNum == 2 => Select the top two RAs by power – RA1 (60 DYM) and RA2 (100 DYM)
//     - Total VP is 200 DYM => RA1 gets 60/200 = 30% and RA2 gets 100/200 = 50% of rewards
//     - With the above header hash, we will pump for 2040061966151279 aDYM in this epoch (pre-calculated)
//     - RA1 gets 612 018 589 845 383 aDYM, RA2 gets 1 020 030 983 075 639 aDYM => Total is 1 632 049 572 921 022 aDYM
//     - Stream.DistributedCoins += 1 632 049 572 921 022 aDYM
//     - Stream.EpochBudgetLeft -= 1 632 049 572 921 022 aDYM
//     - IRO.SoldAmt increases (we don't test the exact number here)
//     - x/streamer balance -= 1 632 049 572 921 022 aDYM
//     - EventPumped and EventBurn occur
//  13. Simulate the next epoch start. It should end the previous epoch and start a new one.
//  14. Validate the Pump Stream:
//     - FilledEpochs += 1
//     - EpochBudget == EpochBudgetLeft == (100 DYM - 1 632 049 572 921 022 aDYM) / 9 == 11 110 929 772 269 675 442
//  15. Settle the IRO
//  16. Put a predictable block hash in sdk.Context that the pump won’t definitely happen
//  17. Simulate a new block and verify that the pump wasn’t executed (see 10)
//  18. Put a predictable block hash in sdk.Context that the pump will definitely happen.
//  19. Simulate a new block and verify that the pump was executed:
//     - TopRollappNum == 2 => Select the top two RAs by power – RA1 (60 DYM) and RA2 (100 DYM)
//     - Total VP is 200 DYM => RA1 gets 60/200 = 30% and RA2 gets 100/200 = 50% of rewards
//     - With the above header hash, we will pump for 2 756 280 626 895 729 aDYM in this epoch (pre-calculated)
//     - RA1 gets 826 884 188 068 718 aDYM, RA2 gets 1 378 140 313 447 864 aDYM => Total is 2 205 024 501 516 582 aDYM
//     - Stream.DistributedCoins += 2 205 024 501 516 582 aDYM
//     - Stream.EpochBudgetLeft -= 2 205 024 501 516 582 aDYM
//     - x/streamer balance -= 2 205 024 501 516 582 aDYM
//     - EventPumped, EvenBurn and swap events occur
func (s *KeeperTestSuite) runPumpStreamTest(tc pumpTestCase) {
	// Step 1: Create 4 rollapps
	rollapps := s.createRollapps(4)

	// Step 2: Create 2 delegators with staking power
	delegators := s.createDelegatorsWithStaking()

	// Step 3: Vote on rollapps with specific distribution
	s.voteOnRollapps(delegators)

	// Step 4: Create IRO
	planID1, iroReserved1 := s.createIRO(rollapps[0])
	planID2, iroReserved2 := s.createIRO(rollapps[1])
	planIDs := []string{planID1, planID2}

	// Step 5: Create Pump Stream
	startTime := time.Now().Add(-time.Minute)
	streamID, _ := s.CreatePumpStream(tc.streamCoins, startTime, tc.epochIdentifier, tc.numEpochsPaidOver, tc.pumpParams)

	// Step 6: Validate initial pump stream state
	s.validateInitialPumpStream(streamID)

	// Step 7: Simulate epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 8: Validate pump stream after epoch start
	s.validatePumpStreamAfterEpochStart(streamID, tc)

	// Step 9: Set predictable hash for no pump
	s.Ctx = hashNoPump(s.Ctx)

	// Step 10: Simulate block and verify no pump
	s.simulateBlockAndVerifyNoPump(s.Ctx, streamID, planIDs)

	// Step 11: Set predictable hash for pump execution
	s.Ctx = hashPump(s.Ctx)

	// Step 12: Simulate block and verify pump execution
	s.simulateBlockAndVerifyPump(s.Ctx, streamID, planIDs)

	// Step 13: Simulate next epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 14: Validate pump stream after second epoch
	s.validatePumpStreamAfterSecondEpoch(streamID)

	// Step 15: Settle IROs
	s.settleIRO(rollapps[0], iroReserved1)
	s.settleIRO(rollapps[1], iroReserved2)

	// Step 16: Set predictable hash for no pump (post-settlement)
	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Ctx = hashNoPump(s.Ctx)

	// Step 17: Simulate block and verify no pump (post-settlement)
	s.simulateBlockAndVerifyNoPumpPostSettlement(s.Ctx, streamID)

	// Step 18: Set predictable hash for pump (post-settlement)
	s.Ctx = hashPump(s.Ctx)

	// Step 19: Simulate block and verify pump with AMM swap (post-settlement)
	s.simulateBlockAndVerifyPumpWithAMM(s.Ctx, streamID)
}

// Helper functions

func (s *KeeperTestSuite) createRollapps(count int) []string {
	rollapps := make([]string, count)
	for i := 0; i < count; i++ {
		rollapps[i] = s.CreateDefaultRollapp()
	}
	return rollapps
}

func (s *KeeperTestSuite) createDelegatorsWithStaking() []sdk.AccAddress {
	// Create validator
	val := s.CreateValidator()
	valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
	s.Require().NoError(err)

	// Create two delegators with 100 DYM each
	del1 := s.CreateDelegator(valAddr, commontypes.DYM.MulRaw(100))
	del2 := s.CreateDelegator(valAddr, commontypes.DYM.MulRaw(100))

	delAddr1, _ := sdk.AccAddressFromBech32(del1.GetDelegatorAddr())
	delAddr2, _ := sdk.AccAddressFromBech32(del2.GetDelegatorAddr())
	return []sdk.AccAddress{delAddr1, delAddr2}
}

func (s *KeeperTestSuite) voteOnRollapps(delegators []sdk.AccAddress) {
	// Get gauge IDs for rollapps
	gauges := s.App.IncentivesKeeper.GetGauges(s.Ctx)

	// User 1 votes: 60 on RA1, 40 on RA2
	vote1 := sponsorshiptypes.MsgVote{
		Voter: delegators[0].String(),
		Weights: []sponsorshiptypes.GaugeWeight{
			{GaugeId: gauges[0].Id, Weight: commontypes.DYM.MulRaw(60)},
			{GaugeId: gauges[1].Id, Weight: commontypes.DYM.MulRaw(40)},
		},
	}
	s.Vote(vote1)

	// User 2 votes: 60 on RA2, 40 on RA3
	vote2 := sponsorshiptypes.MsgVote{
		Voter: delegators[1].String(),
		Weights: []sponsorshiptypes.GaugeWeight{
			{GaugeId: gauges[1].Id, Weight: commontypes.DYM.MulRaw(60)},
			{GaugeId: gauges[2].Id, Weight: commontypes.DYM.MulRaw(40)},
		},
	}
	s.Vote(vote2)
}

func (s *KeeperTestSuite) createIRO(rollappID string) (planID string, reservedAmt math.Int) {
	k := s.App.IROKeeper
	curve := irotypes.DefaultBondingCurve()
	incentives := irotypes.DefaultIncentivePlanParams()
	allocation := math.NewInt(100).MulRaw(1e18)
	liquidityPart := irotypes.DefaultParams().MinLiquidityPart

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappID)
	planID, err := k.CreatePlan(s.Ctx, sdk.DefaultBondDenom, allocation, time.Hour, time.Now(), true, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	return planID, k.MustGetPlan(s.Ctx, planID).SoldAmt
}

func (s *KeeperTestSuite) validateInitialPumpStream(streamID uint64) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)
	s.Require().True(stream.IsPumpStream())

	// Check that only BaseDenom is in coins
	s.Require().Len(stream.Coins, 1)
	s.Require().Equal(sdk.DefaultBondDenom, stream.Coins[0].Denom)

	// EpochBudget and EpochBudgetLeft should be 0 initially
	s.Require().True(stream.PumpParams.EpochBudget.Equal(commontypes.DYM.MulRaw(10)))
	s.Require().True(stream.PumpParams.EpochBudgetLeft.Equal(commontypes.DYM.MulRaw(10)))

	// Stream should not be active yet in epoch terms
	s.Require().Equal(uint64(0), stream.FilledEpochs)
}

func (s *KeeperTestSuite) simulateEpochStart(epochIdentifier string) {
	info := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, epochIdentifier)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(info.Duration).Add(time.Second))
	s.App.EpochsKeeper.BeginBlocker(s.Ctx)
}

func (s *KeeperTestSuite) validatePumpStreamAfterEpochStart(streamID uint64, tc pumpTestCase) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)

	// EpochBudget and EpochBudgetLeft should be calculated based on NumEpochsPaidOver
	expectedBudget := stream.Coins[0].Amount.Quo(math.NewIntFromUint64(tc.numEpochsPaidOver))
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudget)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudgetLeft)
}

func hashNoPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in no pump
	// The value is found experimentally in TestRandom()
	hash := make([]byte, 32)
	hash[31] = 9
	return ctx.WithHeaderHash(hash)
}

func hashPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in a pump
	// The value is found experimentally in TestRandom()
	hash := make([]byte, 32)
	hash[31] = 4
	return ctx.WithHeaderHash(hash)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyNoPump(ctx sdk.Context, streamID uint64, planIDs []string) {
	// Get initial state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	initialDistributedCoins := initialStream.DistributedCoins
	initialEpochBudgetLeft := initialStream.PumpParams.EpochBudgetLeft

	// Get initial IRO sold amounts
	initialSoldAmts := make([]math.Int, len(planIDs))
	for i, planID := range planIDs {
		plan := s.App.IROKeeper.MustGetPlan(ctx, planID)
		initialSoldAmts[i] = plan.SoldAmt
	}

	initialStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)

	// Simulate pump distribution call
	pumpStreams := s.App.StreamerKeeper.GetActiveStreams(ctx)
	var pumpStreamsList []types.Stream
	for _, stream := range pumpStreams {
		if stream.IsPumpStream() {
			pumpStreamsList = append(pumpStreamsList, stream)
		}
	}

	// Execute pump distribution with no-pump context
	err = s.App.StreamerKeeper.DistributePumpStreams(ctx, pumpStreamsList)
	s.Require().NoError(err)

	// Verify no changes occurred
	finalStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	s.Require().True(finalStream.DistributedCoins.Equal(initialDistributedCoins))
	s.Require().Equal(initialEpochBudgetLeft, finalStream.PumpParams.EpochBudgetLeft)

	// Verify IRO sold amounts unchanged
	for i, planID := range planIDs {
		finalPlan := s.App.IROKeeper.MustGetPlan(ctx, planID)
		s.Require().Equal(initialSoldAmts[i], finalPlan.SoldAmt)
	}

	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	s.Require().Equal(initialStreamerBalance, finalStreamerBalance)

	// Verify no EventPumped event was emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 0)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyPump(ctx sdk.Context, streamID uint64, planIDs []string) {
	// Get initial state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	// Get initial IRO sold amounts
	initialSoldAmts := make([]math.Int, len(planIDs))
	for i, planID := range planIDs {
		plan := s.App.IROKeeper.MustGetPlan(ctx, planID)
		initialSoldAmts[i] = plan.SoldAmt
	}

	initialStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)

	// Simulate pump distribution call
	pumpStreams := s.App.StreamerKeeper.GetActiveStreams(ctx)
	var pumpStreamsList []types.Stream
	for _, stream := range pumpStreams {
		if stream.IsPumpStream() {
			pumpStreamsList = append(pumpStreamsList, stream)
		}
	}

	// Execute pump distribution with pump context
	err = s.App.StreamerKeeper.DistributePumpStreams(ctx, pumpStreamsList)
	s.Require().NoError(err)

	// Verify changes occurred
	finalStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	// DistributedCoins should have increased
	distributed := finalStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom)
	expectedDistr, ok := math.NewIntFromString("1632049572921022")
	s.Require().True(ok)
	s.Require().True(distributed.Equal(expectedDistr), "expected %s, got %s", expectedDistr, distributed)

	// EpochBudgetLeft should have decreased
	left := finalStream.PumpParams.EpochBudgetLeft
	expectedLeft := initialStream.PumpParams.EpochBudgetLeft.Sub(expectedDistr)
	s.Require().True(left.Equal(expectedLeft), "expected %s, got %s", expectedLeft, left)

	// EpochBudget should be the same
	budget := finalStream.PumpParams.EpochBudget
	expectedBudget := initialStream.PumpParams.EpochBudget
	s.Require().True(budget.Equal(expectedBudget), "expected %s, got %s", expectedBudget, budget)

	// IRO plan SoldAmt should have changed (increased)
	for i, planID := range planIDs {
		finalPlan := s.App.IROKeeper.MustGetPlan(ctx, planID)
		s.Require().True(finalPlan.SoldAmt.GT(initialSoldAmts[i]))
	}

	// x/streamer balance should have decreased
	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	expectedStreamerBalance := initialStreamerBalance.Amount.Sub(expectedDistr)
	s.Require().Equal(expectedStreamerBalance, finalStreamerBalance.Amount, "expected %s, got %s", expectedStreamerBalance, finalStreamerBalance.Amount)

	// Verify EventPumped and EventBurn events were emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}

func (s *KeeperTestSuite) validatePumpStreamAfterSecondEpoch(streamID uint64) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)

	// FilledEpochs should have increased by 1
	s.Require().Equal(uint64(1), stream.FilledEpochs)

	// EpochBudget and EpochBudgetLeft should be recalculated
	expectedBudget, ok := math.NewIntFromString("11110929772269675442")
	s.Require().True(ok)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudget)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudgetLeft)
}

func (s *KeeperTestSuite) settleIRO(rollappID string, reserveAmt math.Int) {
	plan, found := s.App.IROKeeper.GetPlanByRollapp(s.Ctx, rollappID)
	s.Require().True(found)

	// Fund module with insufficient funds for settlement
	// Sold amount
	iroDenom := plan.TotalAllocation.Denom
	amt := plan.SoldAmt.Sub(reserveAmt)
	s.FundModuleAcc(irotypes.ModuleName, sdk.NewCoins(sdk.NewCoin(iroDenom, amt)))

	// Settlement token
	rollappDenom := fmt.Sprintf("hui/%s", rollappID)
	amt = plan.TotalAllocation.Amount
	s.FundModuleAcc(irotypes.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))

	err := s.App.IROKeeper.Settle(s.Ctx, rollappID, rollappDenom)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyNoPumpPostSettlement(ctx sdk.Context, streamID uint64) {
	// Similar to simulateBlockAndVerifyNoPump but for post-settlement state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	// Execute pump distribution
	pumpStreams := s.App.StreamerKeeper.GetActiveStreams(ctx)
	var pumpStreamsList []types.Stream
	for _, stream := range pumpStreams {
		if stream.IsPumpStream() {
			pumpStreamsList = append(pumpStreamsList, stream)
		}
	}

	err = s.App.StreamerKeeper.DistributePumpStreams(ctx, pumpStreamsList)
	s.Require().NoError(err)

	// Verify no changes
	finalStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	s.Require().True(finalStream.DistributedCoins.Equal(initialStream.DistributedCoins))
	s.Require().Equal(initialStream.PumpParams.EpochBudgetLeft, finalStream.PumpParams.EpochBudgetLeft)

	// Verify no events
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 0)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyPumpWithAMM(ctx sdk.Context, streamID uint64) {
	// Similar to simulateBlockAndVerifyPump but expects AMM swap events post-settlement
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	// Get initial streamer balance
	initialStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)

	// Execute pump distribution
	pumpStreams := s.App.StreamerKeeper.GetActiveStreams(ctx)
	var pumpStreamsList []types.Stream
	for _, stream := range pumpStreams {
		if stream.IsPumpStream() {
			pumpStreamsList = append(pumpStreamsList, stream)
		}
	}

	err = s.App.StreamerKeeper.DistributePumpStreams(ctx, pumpStreamsList)
	s.Require().NoError(err)

	// Verify changes
	finalStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	// DistributedCoins should have increased
	distributed, ok := math.NewIntFromString("2205024501516582")
	s.Require().True(ok)
	finalDistributed := finalStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom)
	expectedDistr := initialStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom).Add(distributed)
	s.Require().True(finalDistributed.Equal(expectedDistr), "expected %s, got %s", expectedDistr, finalDistributed)

	// EpochBudgetLeft should have decreased
	left := finalStream.PumpParams.EpochBudgetLeft
	expectedLeft := initialStream.PumpParams.EpochBudgetLeft.Sub(distributed)
	s.Require().True(left.Equal(expectedLeft), "expected %s, got %s", expectedLeft, left)

	// EpochBudget should be the same
	budget := finalStream.PumpParams.EpochBudget
	expectedBudget := initialStream.PumpParams.EpochBudget
	s.Require().True(budget.Equal(expectedBudget), "expected %s, got %s", expectedBudget, budget)

	// x/streamer balance should have decreased by expected amount (21902 aDYM according to comment)
	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	expectedStreamerBalance := initialStreamerBalance.Amount.Sub(distributed)
	s.Require().Equal(expectedStreamerBalance, finalStreamerBalance.Amount, "expected %s, got %s", expectedStreamerBalance, finalStreamerBalance.Amount)

	// Verify pump events, burn events and swap events
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}
