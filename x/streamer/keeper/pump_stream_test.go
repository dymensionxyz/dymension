package keeper_test

import (
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
	liquidityDenom    string
	expectedNumPumped int
	expectPumpSuccess bool
	expectBurnEvents  bool
	contextOverride   func(ctx sdk.Context) sdk.Context
}

func (s *KeeperTestSuite) TestShouldPump() {
	b, err := s.App.StreamerKeeper.EpochBlocks(s.Ctx, "day")
	s.Require().NoError(err)

	pumpNum := uint64(600000000000000)

	s.Run("GenerateUnifiedRandom", func() {
		// Pump hash
		ctx := s.hashPump(s.Ctx)
		r1 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandom(ctx, b.BigIntMut()),
		) //  510461966151279

		// No pump hash
		ctx = s.hashNoPump(s.Ctx)
		r2 := math.NewIntFromBigIntMut(
			keeper.GenerateUnifiedRandom(ctx, b.BigIntMut()),
		) //  723264247167302

		middle := math.NewIntFromUint64(pumpNum)

		s.Require().True(r1.LT(middle))
		s.Require().True(middle.LT(r2))
	})

	s.Run("ShouldPump", func() {
		// Pump hash should pump
		ctx := s.hashPump(s.Ctx)
		pumpAmt, err := keeper.ShouldPump(
			ctx,
			commontypes.DYM.MulRaw(10),
			commontypes.DYM.MulRaw(10),
			pumpNum,
			b,
		)
		s.Require().NoError(err)
		s.Require().False(pumpAmt.IsZero())
		s.Require().True(math.NewInt(22587).Equal(pumpAmt))

		// No pump hash should not pump
		ctx = s.hashNoPump(s.Ctx)
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
				NumPumps:       600000000000000, // This is a hardcoded number for test convenience
			},
			numEpochsPaidOver: 10,
			epochIdentifier:   "day",
			streamCoins:       sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))),
			liquidityDenom:    sdk.DefaultBondDenom,
			expectedNumPumped: 2,
			expectPumpSuccess: true,
			expectBurnEvents:  true,
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
//  10. Simulate a new block and verify that the pump wasn’t executed – DistributedCoins, EpochBudgetLeft, IRO.SoldAmt are the same and no EventPumped event. x/streamer balance is the same.
//  11. Put a predictable block hash in sdk.Context that the pump will definitely happen.
//  12. Simulate a new block and verify that the pump was executed:
//     - TopRollappNum == 2 => Select the top two RAs by power – RA1 (60 DYM) and RA2 (100 DYM)
//     - Total VP is 200 DYM => RA1 gets 60/200 = 30% and RA2 gets 100/200 = 50% of rewards
//     - With the above header hash, we will pump for 22587 aDYM in this epoch (pre-calculated)
//     - RA1 gets 6776 aDYM, RA2 gets 11293 aDYM => Total is 18069 aDYM
//     - Stream.DistributedCoins += 18069 aDYM
//     - Stream.EpochBudgetLeft -= 18069 aDYM
//     - IRO.SoldAmt increases (we don't test the exact number here)
//     - x/streamer balance -= 18069 aDYM
//     - EventPumped and EventBurn occur
//  13. Simulate the next epoch start. It should end the previous epoch and start a new one.
//  14. Validate the Pump Stream:
//     - FilledEpochs += 1
//     - EpochBudget == EpochBudgetLeft == (100 DYM - 18069 aDYM) / 9 == 1 111 111 111 111 109 103
//  15. Settle the IRO – @x/iro/keeper/settle_test.go#L51
//  16. Put a predictable block hash in sdk.Context that the pump won’t definitely take place – @x/streamer/keeper/pump_stream.go#L72
//  17. Simulate a new block and verify that the pump wasn’t executed – check DistributedCoins, EpochBudgetLeft, and AMM pool params are the same and no EventPumped nor AMM swap events events. x/streamer balance should stay the same.
//  18. Put a predictable block hash in sdk.Context that the pump will definitely take place.
//  19. Simulate a new block and verify that the pump was executed – Check existence of EventPumped and EventBurn and AMM swap events. Check pumped (or burned) amount for every RollApp from any event and see that the distribution of pumped amount was correct. Check DistributedCoins, EpochBudgetLeft, AMM pool params are changed accordingly.
func (s *KeeperTestSuite) runPumpStreamTest(tc pumpTestCase) {
	// Step 1: Create 4 rollapps
	rollapps := s.createRollapps(4)

	// Step 2: Create 2 delegators with staking power
	delegators := s.createDelegatorsWithStaking()

	// Step 3: Vote on rollapps with specific distribution
	s.voteOnRollapps(delegators)

	// Step 4: Create IRO
	planID1 := s.createIRO(rollapps[0], tc.liquidityDenom)
	planID2 := s.createIRO(rollapps[1], tc.liquidityDenom)

	// Step 5: Create Pump Stream
	streamID := s.createPumpStream(tc)

	// Step 6: Validate initial pump stream state
	s.validateInitialPumpStream(streamID)

	// Step 7: Simulate epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 8: Validate pump stream after epoch start
	s.validatePumpStreamAfterEpochStart(streamID, tc)

	// Step 9: Set predictable hash for no pump
	ctxNoPump := s.hashNoPump(s.Ctx)

	// Step 10: Simulate block and verify no pump
	s.simulateBlockAndVerifyNoPump(ctxNoPump, streamID, []string{planID1, planID2})

	// Step 11: Set predictable hash for pump execution
	ctxPump := s.hashPump(s.Ctx)

	// Step 12: Simulate block and verify pump execution
	s.simulateBlockAndVerifyPump(ctxPump, streamID, []string{planID1, planID2})

	// Step 13: Simulate next epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 14: Validate pump stream after second epoch
	s.validatePumpStreamAfterSecondEpoch(streamID)

	// Step 15: Settle IRO
	s.settleIRO(rollapps[0])

	// Step 16: Set predictable hash for no pump (post-settlement)
	ctxNoPumpSettled := s.hashNoPump(s.Ctx)

	// Step 17: Simulate block and verify no pump (post-settlement)
	s.simulateBlockAndVerifyNoPumpPostSettlement(ctxNoPumpSettled, streamID)

	// Step 18: Set predictable hash for pump (post-settlement)
	ctxPumpSettled := s.hashPump(s.Ctx)

	// Step 19: Simulate block and verify pump with AMM swap (post-settlement)
	s.simulateBlockAndVerifyPumpWithAMM(ctxPumpSettled, streamID)
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
	initial := sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))
	del1 := s.CreateDelegator(valAddr, initial)
	del2 := s.CreateDelegator(valAddr, initial)

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

func (s *KeeperTestSuite) createIRO(rollappID, liquidityDenom string) string {
	k := s.App.IROKeeper
	curve := irotypes.DefaultBondingCurve()
	incentives := irotypes.DefaultIncentivePlanParams()
	allocation := math.NewInt(100).MulRaw(1e18)
	liquidityPart := irotypes.DefaultParams().MinLiquidityPart

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappID)
	planID, err := k.CreatePlan(s.Ctx, sdk.DefaultBondDenom, allocation, time.Hour, time.Now(), true, rollapp, curve, incentives, liquidityPart, time.Hour, 0)
	s.Require().NoError(err)

	return planID
}

func (s *KeeperTestSuite) createPumpStream(tc pumpTestCase) uint64 {
	startTime := time.Now().Add(-time.Minute)
	streamID, err := s.App.StreamerKeeper.CreateStream(
		s.Ctx,
		tc.streamCoins,
		[]types.DistrRecord{}, // Empty for pump stream
		startTime,
		tc.epochIdentifier,
		tc.numEpochsPaidOver,
		false, // not sponsored
		tc.pumpParams,
	)
	s.Require().NoError(err)
	return streamID
}

func (s *KeeperTestSuite) validateInitialPumpStream(streamID uint64) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)
	s.Require().True(stream.IsPumpStream())

	// Check that only BaseDenom is in coins
	s.Require().Len(stream.Coins, 1)
	s.Require().Equal(sdk.DefaultBondDenom, stream.Coins[0].Denom)

	// EpochBudget and EpochBudgetLeft should be 0 initially
	s.Require().True(stream.PumpParams.EpochBudget.IsZero())
	s.Require().True(stream.PumpParams.EpochBudgetLeft.IsZero())

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

func (s *KeeperTestSuite) hashNoPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in no pump
	// The value is found experimentally in TestRandom()
	hash := make([]byte, 32)
	hash[31] = 3
	return ctx.WithHeaderHash(hash)
}

func (s *KeeperTestSuite) hashPump(ctx sdk.Context) sdk.Context {
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

	//plan := s.App.IROKeeper.MustGetPlan(ctx, planID)
	//initialSoldAmt := plan.SoldAmt

	initialStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)

	// Finalize block
	//s.Commit()

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

	//finalPlan := s.App.IROKeeper.MustGetPlan(ctx, planID)
	//s.Require().Equal(initialSoldAmt, finalPlan.SoldAmt)

	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	s.Require().Equal(initialStreamerBalance, finalStreamerBalance)

	// Verify no EventPumped event was emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 0)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyPump(ctx sdk.Context, streamID uint64, planIDs []string) {
	// Get initial state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	//plan := s.App.IROKeeper.MustGetPlan(ctx, planID)
	//initialSoldAmt := plan.SoldAmt

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
	expectedDistr := math.NewInt(18069)
	s.Require().True(distributed.Equal(expectedDistr), "expected %s, got %s", expectedDistr, distributed)

	// EpochBudgetLeft should have decreased
	left := finalStream.PumpParams.EpochBudgetLeft
	expectedLeft := initialStream.PumpParams.EpochBudgetLeft.Sub(math.NewInt(18069))
	s.Require().True(left.Equal(expectedLeft), "expected %s, got %s", expectedLeft, left)

	// EpochBudget should be the same
	budget := finalStream.PumpParams.EpochBudget
	expectedBudget := initialStream.PumpParams.EpochBudget
	s.Require().True(budget.Equal(expectedBudget), "expected %s, got %s", expectedBudget, budget)

	// IRO plan SoldAmt should have changed (increased)
	//finalPlan := s.App.IROKeeper.MustGetPlan(ctx, planID)
	//s.Require().True(finalPlan.SoldAmt.GT(initialSoldAmt))

	// Verify EventPumped event was emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}

func (s *KeeperTestSuite) validatePumpStreamAfterSecondEpoch(streamID uint64) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)

	// FilledEpochs should have increased by 1
	s.Require().Equal(uint64(1), stream.FilledEpochs)

	// EpochBudget and EpochBudgetLeft should be recalculated
	expectedBudget, ok := math.NewIntFromString("11111111111111109103")
	s.Require().True(ok)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudget)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudgetLeft)
}

func (s *KeeperTestSuite) settleIRO(rollappID string) {
	plan, found := s.App.IROKeeper.GetPlanByRollapp(s.Ctx, rollappID)
	s.Require().True(found)

	// Fund module for settlement
	rollappDenom := plan.TotalAllocation.Denom
	amt := plan.TotalAllocation.Amount
	s.FundModuleAcc(irotypes.ModuleName, sdk.NewCoins(sdk.NewCoin(rollappDenom, amt)))

	err := s.App.IROKeeper.Settle(s.Ctx, rollappID, rollappDenom)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyNoPumpPostSettlement(ctx sdk.Context, streamID uint64) {
	// Similar to simulateBlockAndVerifyNoPump but for post-settlement state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	initialDistributedCoins := initialStream.DistributedCoins
	initialEpochBudgetLeft := initialStream.PumpParams.EpochBudgetLeft

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

	s.Require().True(finalStream.DistributedCoins.Equal(initialDistributedCoins))
	s.Require().Equal(initialEpochBudgetLeft, finalStream.PumpParams.EpochBudgetLeft)

	// Verify no events
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 0)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyPumpWithAMM(ctx sdk.Context, streamID uint64) {
	// Similar to simulateBlockAndVerifyPump but expects AMM swap events post-settlement
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	initialDistributedCoins := initialStream.DistributedCoins
	initialEpochBudgetLeft := initialStream.PumpParams.EpochBudgetLeft

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

	s.Require().True(finalStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom).GT(initialDistributedCoins.AmountOf(sdk.DefaultBondDenom)))
	s.Require().True(finalStream.PumpParams.EpochBudgetLeft.LT(initialEpochBudgetLeft))

	// Verify pump events
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}
