package keeper_test

import (
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
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

func (s *KeeperTestSuite) TestPumpStreamComprehensive() {
	// Define test cases with different parameters
	testCases := []pumpTestCase{
		{
			name: "basic pump stream scenario",
			pumpParams: &types.MsgCreateStream_PumpParams{
				NumTopRollapps: 2,
				NumPumps:       10,
			},
			numEpochsPaidOver: 10,
			epochIdentifier:   "day",
			streamCoins:       sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(100))),
			liquidityDenom:    sdk.DefaultBondDenom,
			expectedNumPumped: 2,
			expectPumpSuccess: true,
			expectBurnEvents:  true,
		},
		{
			name: "high frequency pumps",
			pumpParams: &types.MsgCreateStream_PumpParams{
				NumTopRollapps: 3,
				NumPumps:       50,
			},
			numEpochsPaidOver: 5,
			epochIdentifier:   "day",
			streamCoins:       sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(50))),
			liquidityDenom:    sdk.DefaultBondDenom,
			expectedNumPumped: 3,
			expectPumpSuccess: true,
			expectBurnEvents:  true,
		},
		{
			name: "single rollapp target",
			pumpParams: &types.MsgCreateStream_PumpParams{
				NumTopRollapps: 1,
				NumPumps:       5,
			},
			numEpochsPaidOver: 20,
			epochIdentifier:   "day",
			streamCoins:       sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(200))),
			liquidityDenom:    sdk.DefaultBondDenom,
			expectedNumPumped: 1,
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

func (s *KeeperTestSuite) runPumpStreamTest(tc pumpTestCase) {
	// Step 1: Create 4 rollapps
	rollapps := s.createRollapps(4)

	// Step 2: Create 2 delegators with staking power
	delegators := s.createDelegatorsWithStaking()

	// Step 3: Vote on rollapps with specific distribution
	s.voteOnRollapps(delegators, rollapps)

	// Step 4: Create IRO
	planID := s.createIRO(rollapps[0], tc.liquidityDenom)

	// Step 5: Create Pump Stream
	streamID := s.createPumpStream(tc)

	// Step 6: Validate initial pump stream state
	s.validateInitialPumpStream(streamID)

	// Step 7: Simulate epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 8: Validate pump stream after epoch start
	s.validatePumpStreamAfterEpochStart(streamID, tc)

	// Step 9: Set predictable hash for no pump
	ctxNoPump := s.setPredictableHashForNoPump(s.Ctx)

	// Step 10: Simulate block and verify no pump
	s.simulateBlockAndVerifyNoPump(ctxNoPump, streamID, planID)

	// Step 11: Set predictable hash for pump execution
	ctxPump := s.setPredictableHashForPump(s.Ctx)

	// Step 12: Simulate block and verify pump execution
	s.simulateBlockAndVerifyPump(ctxPump, streamID, planID)

	// Step 13: Simulate next epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 14: Validate pump stream after second epoch
	s.validatePumpStreamAfterSecondEpoch(streamID, tc)

	// Step 15: Settle IRO
	s.settleIRO(rollapps[0])

	// Step 16: Set predictable hash for no pump (post-settlement)
	ctxNoPumpSettled := s.setPredictableHashForNoPump(s.Ctx)

	// Step 17: Simulate block and verify no pump (post-settlement)
	s.simulateBlockAndVerifyNoPumpPostSettlement(ctxNoPumpSettled, streamID)

	// Step 18: Set predictable hash for pump (post-settlement)
	ctxPumpSettled := s.setPredictableHashForPump(s.Ctx)

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

func (s *KeeperTestSuite) voteOnRollapps(delegators []sdk.AccAddress, rollapps []string) {
	// Create rollapp gauges
	for _, rollappID := range rollapps {
		_, err := s.App.IncentivesKeeper.CreateRollappGauge(s.Ctx, rollappID)
		s.Require().NoError(err)
	}

	// Get gauge IDs for rollapps
	gauges := s.App.IncentivesKeeper.GetGauges(s.Ctx)
	rollappGauges := make([]uint64, 0)
	for _, gauge := range gauges {
		if gauge.GetRollapp() != nil {
			rollappGauges = append(rollappGauges, gauge.Id)
		}
	}

	// User 1 votes: 60 on RA1, 40 on RA2
	if len(rollappGauges) >= 2 {
		vote1 := sponsorshiptypes.MsgVote{
			Voter: delegators[0].String(),
			Weights: []sponsorshiptypes.GaugeWeight{
				{GaugeId: rollappGauges[0], Weight: commontypes.DYM.MulRaw(60)},
				{GaugeId: rollappGauges[1], Weight: commontypes.DYM.MulRaw(40)},
			},
		}
		s.vote(vote1)
	}

	// User 2 votes: 60 on RA2, 40 on RA3
	if len(rollappGauges) >= 3 {
		vote2 := sponsorshiptypes.MsgVote{
			Voter: delegators[1].String(),
			Weights: []sponsorshiptypes.GaugeWeight{
				{GaugeId: rollappGauges[1], Weight: commontypes.DYM.MulRaw(60)},
				{GaugeId: rollappGauges[2], Weight: commontypes.DYM.MulRaw(40)},
			},
		}
		s.vote(vote2)
	}
}

func (s *KeeperTestSuite) createIRO(rollappID, liquidityDenom string) uint64 {
	k := s.App.IROKeeper
	amt := math.NewInt(1_000_000).MulRaw(1e18)
	startTime := time.Now()

	rollapp, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappID)
	planID, err := k.CreatePlan(s.Ctx, liquidityDenom, amt, time.Hour, startTime, true, rollapp, irotypes.BondingCurve{}, irotypes.IncentivePlanParams{}, math.LegacyOneDec(), time.Hour, 0)
	s.Require().NoError(err)

	planIDUint, err := strconv.ParseUint(planID, 10, 64)
	s.Require().NoError(err)
	return planIDUint
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
	s.Ctx = s.Ctx.WithBlockTime(info.CurrentEpochStartTime.Add(info.Duration).Add(time.Minute))
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

func (s *KeeperTestSuite) setPredictableHashForNoPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in no pump
	// Based on the ShouldPump logic, we want a hash that when modded will be >= numPumps
	hash := make([]byte, 32)
	hash[31] = 255 // High value to ensure no pump
	return ctx.WithHeaderHash(hash)
}

func (s *KeeperTestSuite) setPredictableHashForPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in a pump
	// Low value to ensure pump happens
	hash := make([]byte, 32)
	hash[31] = 1 // Low value to ensure pump
	return ctx.WithHeaderHash(hash)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyNoPump(ctx sdk.Context, streamID, planID uint64) {
	// Get initial state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	initialDistributedCoins := initialStream.DistributedCoins
	initialEpochBudgetLeft := initialStream.PumpParams.EpochBudgetLeft

	plan := s.App.IROKeeper.MustGetPlan(ctx, strconv.FormatUint(planID, 10))
	initialSoldAmt := plan.SoldAmt

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

	finalPlan := s.App.IROKeeper.MustGetPlan(ctx, strconv.FormatUint(planID, 10))
	s.Require().Equal(initialSoldAmt, finalPlan.SoldAmt)

	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	s.Require().Equal(initialStreamerBalance, finalStreamerBalance)

	// Verify no EventPumped event was emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 0)
}

func (s *KeeperTestSuite) simulateBlockAndVerifyPump(ctx sdk.Context, streamID, planID uint64) {
	// Get initial state
	initialStream, err := s.App.StreamerKeeper.GetStreamByID(ctx, streamID)
	s.Require().NoError(err)

	initialDistributedCoins := initialStream.DistributedCoins
	initialEpochBudgetLeft := initialStream.PumpParams.EpochBudgetLeft

	plan := s.App.IROKeeper.MustGetPlan(ctx, strconv.FormatUint(planID, 10))
	initialSoldAmt := plan.SoldAmt

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
	s.Require().True(finalStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom).GT(initialDistributedCoins.AmountOf(sdk.DefaultBondDenom)))

	// EpochBudgetLeft should have decreased
	s.Require().True(finalStream.PumpParams.EpochBudgetLeft.LT(initialEpochBudgetLeft))

	// IRO plan SoldAmt should have changed (increased)
	finalPlan := s.App.IROKeeper.MustGetPlan(ctx, strconv.FormatUint(planID, 10))
	s.Require().True(finalPlan.SoldAmt.GT(initialSoldAmt))

	// Verify EventPumped event was emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}

func (s *KeeperTestSuite) validatePumpStreamAfterSecondEpoch(streamID uint64, tc pumpTestCase) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)

	// FilledEpochs should have increased by 1
	s.Require().Equal(uint64(1), stream.FilledEpochs)

	// EpochBudget and EpochBudgetLeft should be recalculated
	expectedBudget := stream.Coins[0].Amount.Quo(math.NewIntFromUint64(tc.numEpochsPaidOver))
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
