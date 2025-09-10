package keeper_test

import (
	"slices"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/utils/rand"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

type pumpTestCase struct {
	pumpParams            *types.MsgCreateStream_PumpParams
	numEpochsPaidOver     uint64
	epochIdentifier       string
	streamCoins           sdk.Coins
	balanceChangeIter1    math.Int
	balanceChangeIter2    math.Int
	epochBudgetAfterIter1 math.Int
}

// Scenario (numbers are just for reference, real numbers are pre-calculated in the test):
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
func (s *KeeperTestSuite) TestPumpStream() {
	// Stop 0: Pre-calculate all important numbers and prepare the test
	tc := s.prepareTestCase()

	// Step 1: Create 4 rollapps
	rollapps := s.createRollapps(4)

	// Step 2: Create 2 delegators with staking power
	delegators := s.createDelegatorsWithStaking()

	// Step 3: Vote on rollapps with specific distribution
	s.voteOnRollapps(delegators)

	// Step 4: Create IRO
	planID1 := s.CreateDefaultPlan(rollapps[0])
	iroReserved1 := s.App.IROKeeper.MustGetPlan(s.Ctx, planID1).SoldAmt
	planID2 := s.CreateDefaultPlan(rollapps[1])
	iroReserved2 := s.App.IROKeeper.MustGetPlan(s.Ctx, planID2).SoldAmt
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
	s.simulateBlockAndVerifyPump(s.Ctx, streamID, planIDs, tc.balanceChangeIter1)

	// Step 13: Simulate next epoch start
	s.simulateEpochStart(tc.epochIdentifier)

	// Step 14: Validate pump stream after second epoch
	s.validatePumpStreamAfterSecondEpoch(streamID, tc.epochBudgetAfterIter1)

	// Step 15: Settle IROs
	s.SettleIRO(rollapps[0], iroReserved1)
	s.SettleIRO(rollapps[1], iroReserved2)

	// Step 16: Set predictable hash for no pump (post-settlement)
	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.Ctx = hashNoPump(s.Ctx)

	// Step 17: Simulate block and verify no pump (post-settlement)
	s.simulateBlockAndVerifyNoPumpPostSettlement(s.Ctx, streamID)

	// Step 18: Set predictable hash for pump (post-settlement)
	s.Ctx = hashPump(s.Ctx)

	// Step 19: Simulate block and verify pump with AMM swap (post-settlement)
	s.simulateBlockAndVerifyPumpWithAMM(s.Ctx, streamID, tc.balanceChangeIter2)
}

func (s *KeeperTestSuite) prepareTestCase() pumpTestCase {
	var (
		epochID               = "day"
		numEpochsPaidOver     = uint64(10)
		remainEpochs          = numEpochsPaidOver
		streamCoinsAmtInitial = commontypes.DYM.MulRaw(100)
		pumpNum               = uint64(7000)
		ctx                   = hashPump(s.Ctx)
		epochBudget           = streamCoinsAmtInitial.Quo(math.NewIntFromUint64(remainEpochs))
		epochBudgetLeft       = epochBudget
		numTopRollapps        = uint32(2)
	)

	b, err := s.App.StreamerKeeper.EpochBlocks(s.Ctx, epochID)
	s.Require().NoError(err)

	// Pump amount on step (12)
	pumpAmtIter1, err := keeper.ShouldPump(ctx, types.PumpParams{
		NumTopRollapps:  numTopRollapps,
		EpochBudget:     epochBudget,
		EpochBudgetLeft: epochBudgetLeft,
		NumPumps:        pumpNum,
		PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
	}, b)
	s.Require().NoError(err)
	s.Require().True(!pumpAmtIter1.IsZero())

	remainEpochs = remainEpochs - 1
	var (
		// RA1 gets 30%
		// RA2 gets 50%
		// RA3 gets 20% but is not selected to pump
		// => Normalize:
		// RA1 gets 3/8 = 37.5%
		// RA2 gets 5/8 = 62.5%
		ra1Share    = pumpAmtIter1.MulRaw(3).QuoRaw(8) // 37.5%
		ra2Share    = pumpAmtIter1.MulRaw(5).QuoRaw(8) // 62.5%
		changeIter1 = ra1Share.Add(ra2Share)

		streamCoinsAmtAfterPump  = streamCoinsAmtInitial.Sub(changeIter1)
		epochBudgetAfterPump     = streamCoinsAmtAfterPump.Quo(math.NewIntFromUint64(remainEpochs))
		epochBudgetLeftAfterPump = epochBudgetAfterPump
	)

	// Pump amount on step (19)
	pumpAmtIter2, err := keeper.ShouldPump(ctx, types.PumpParams{
		NumTopRollapps:  numTopRollapps,
		EpochBudget:     epochBudgetAfterPump,
		EpochBudgetLeft: epochBudgetLeftAfterPump,
		NumPumps:        pumpNum,
		PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
	}, b)
	s.Require().NoError(err)
	s.Require().True(!pumpAmtIter2.IsZero())
	ra1Share = pumpAmtIter2.MulRaw(3).QuoRaw(8) // 37.5%
	ra2Share = pumpAmtIter2.MulRaw(5).QuoRaw(8) // 62.5%
	changeIter2 := ra1Share.Add(ra2Share)

	return pumpTestCase{
		pumpParams: &types.MsgCreateStream_PumpParams{
			NumTopRollapps: numTopRollapps,
			NumPumps:       pumpNum,
			PumpDistr:      types.PumpDistr_PUMP_DISTR_UNIFORM,
		},
		numEpochsPaidOver:     numEpochsPaidOver,
		epochIdentifier:       epochID,
		streamCoins:           sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, streamCoinsAmtInitial)),
		balanceChangeIter1:    changeIter1,
		balanceChangeIter2:    changeIter2,
		epochBudgetAfterIter1: epochBudgetLeftAfterPump,
	}
}

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
	hash[31] = 5
	headerInfo := ctx.HeaderInfo()
	headerInfo.Hash = hash
	return ctx.WithHeaderInfo(headerInfo)
}

func hashPump(ctx sdk.Context) sdk.Context {
	// Create a header hash that will result in a pump
	// The value is found experimentally in TestRandom()
	hash := make([]byte, 32)
	hash[31] = 9
	headerInfo := ctx.HeaderInfo()
	headerInfo.Hash = hash
	return ctx.WithHeaderInfo(headerInfo)
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

func (s *KeeperTestSuite) simulateBlockAndVerifyPump(ctx sdk.Context, streamID uint64, planIDs []string, expectedChange math.Int) {
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
	s.Require().True(distributed.Equal(expectedChange), "expected %s, got %s", expectedChange, distributed)

	// EpochBudgetLeft should have decreased
	left := finalStream.PumpParams.EpochBudgetLeft
	expectedLeft := initialStream.PumpParams.EpochBudgetLeft.Sub(expectedChange)
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
	expectedStreamerBalance := initialStreamerBalance.Amount.Sub(expectedChange)
	s.Require().Equal(expectedStreamerBalance, finalStreamerBalance.Amount, "expected %s, got %s", expectedStreamerBalance, finalStreamerBalance.Amount)

	// Verify EventPumped and EventBurn events were emitted
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}

func (s *KeeperTestSuite) validatePumpStreamAfterSecondEpoch(streamID uint64, expectedBudget math.Int) {
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)

	// FilledEpochs should have increased by 1
	s.Require().Equal(uint64(1), stream.FilledEpochs)

	// EpochBudget and EpochBudgetLeft should be recalculated
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudget)
	s.Require().Equal(expectedBudget, stream.PumpParams.EpochBudgetLeft)
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

func (s *KeeperTestSuite) simulateBlockAndVerifyPumpWithAMM(ctx sdk.Context, streamID uint64, expectedChange math.Int) {
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
	finalDistributed := finalStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom)
	expectedDistr := initialStream.DistributedCoins.AmountOf(sdk.DefaultBondDenom).Add(expectedChange)
	s.Require().True(finalDistributed.Equal(expectedDistr), "expected %s, got %s", expectedDistr, finalDistributed)

	// EpochBudgetLeft should have decreased
	left := finalStream.PumpParams.EpochBudgetLeft
	expectedLeft := initialStream.PumpParams.EpochBudgetLeft.Sub(expectedChange)
	s.Require().True(left.Equal(expectedLeft), "expected %s, got %s", expectedLeft, left)

	// EpochBudget should be the same
	budget := finalStream.PumpParams.EpochBudget
	expectedBudget := initialStream.PumpParams.EpochBudget
	s.Require().True(budget.Equal(expectedBudget), "expected %s, got %s", expectedBudget, budget)

	// x/streamer balance should have decreased by expected amount
	finalStreamerBalance := s.App.BankKeeper.GetBalance(ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), sdk.DefaultBondDenom)
	expectedStreamerBalance := initialStreamerBalance.Amount.Sub(expectedChange)
	s.Require().Equal(expectedStreamerBalance, finalStreamerBalance.Amount, "expected %s, got %s", expectedStreamerBalance, finalStreamerBalance.Amount)

	// Verify pump events, burn events and swap events
	s.AssertEventEmitted(ctx, "dymensionxyz.dymension.streamer.EventPumped", 1)
}

func (s *KeeperTestSuite) TestShouldPump() {
	b, err := s.App.StreamerKeeper.EpochBlocks(s.Ctx, "day")
	s.Require().NoError(err)

	pumpNum := uint64(7000)

	s.Run("GenerateUniformRandom", func() {
		// Pump hash
		ctx := hashPump(s.Ctx)
		r1 := math.NewIntFromBigIntMut(
			rand.GenerateUniformRandomMod(ctx, b.BigIntMut()),
		) //  7639

		// No pump hash
		ctx = hashNoPump(s.Ctx)
		r2 := math.NewIntFromBigIntMut(
			rand.GenerateUniformRandomMod(ctx, b.BigIntMut()),
		) //  11118

		middle := math.NewIntFromUint64(pumpNum)

		s.Require().True(r1.LT(middle), "expected r1 < middle, got: %s < %s", r1, middle)
		s.Require().True(middle.LT(r2), "expected middle < r2, got: %s < %s ", middle, r2)
	})

	s.Run("ShouldPump", func() {
		// Pump hash should pump
		ctx := hashPump(s.Ctx)
		pumpAmt, err := keeper.ShouldPump(
			ctx,
			types.PumpParams{
				NumTopRollapps:  0,
				EpochBudget:     commontypes.DYM.MulRaw(10),
				EpochBudgetLeft: commontypes.DYM.MulRaw(10),
				NumPumps:        pumpNum,
				PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
			},
			b,
		)
		s.Require().NoError(err)
		s.Require().False(pumpAmt.IsZero())

		// No pump hash should not pump
		ctx = hashNoPump(s.Ctx)
		pumpAmt, err = keeper.ShouldPump(
			ctx,
			types.PumpParams{
				NumTopRollapps:  0,
				EpochBudget:     commontypes.DYM.MulRaw(10),
				EpochBudgetLeft: commontypes.DYM.MulRaw(10),
				NumPumps:        pumpNum,
				PumpDistr:       types.PumpDistr_PUMP_DISTR_UNIFORM,
			},
			b,
		)
		s.Require().NoError(err)
		s.Require().True(pumpAmt.IsZero())
	})
}

func (s *KeeperTestSuite) TestPumpAmtSamplesUniform() {
	s.T().Skip("This test is for debugging and visualizing the distribution.")

	var (
		epochBudget     = math.NewInt(200_000)
		epochBudgetLeft = epochBudget
		pumpNum         = int64(200)
		ctx             = hashPump(s.Ctx)
		pumpFunc        = types.PumpDistr_PUMP_DISTR_EXPONENTIAL
	)

	values := make([]math.Int, 0, pumpNum)
	total := math.ZeroInt()

	for iteration := int64(0); iteration < pumpNum; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := rand.NextPermutation([32]byte(hash), int(iteration))
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		pumpAmt, err := keeper.PumpAmt(ctx, types.PumpParams{
			NumTopRollapps:  0,
			EpochBudget:     epochBudget,
			EpochBudgetLeft: epochBudgetLeft,
			NumPumps:        uint64(pumpNum),
			PumpDistr:       pumpFunc,
		})
		s.Require().NoError(err)

		epochBudgetLeft = epochBudgetLeft.Sub(pumpAmt)
		total = total.Add(pumpAmt)
		values = append(values, pumpAmt)
	}

	valuesCpy := make([]math.Int, len(values))
	copy(valuesCpy, values)
	slices.SortFunc(values, func(a, b math.Int) int {
		if a.LT(b) {
			return -1
		}
		if a.GT(b) {
			return 1
		}
		return 0
	})

	s.T().Log("Sorted samples – CDF function")
	for _, v := range values {
		println(v.String())
	}

	s.T().Log("Not sorted samples")
	for _, v := range valuesCpy {
		println(v.String())
	}

	s.T().Log("Target mean", epochBudget.QuoRaw(pumpNum))
	s.T().Log("Actual mean", total.QuoRaw(pumpNum))
	s.T().Log("Total distr", total)
}

func (s *KeeperTestSuite) TestExecutePump() {
	testCases := []struct {
		name           string
		pumpAmt        math.Int
		initialBuyAmt  math.Int
		planAllocation math.Int
		preGraduated   bool
		graduated      bool
		settled        bool
	}{
		{
			name:           "pre-graduation buy",
			pumpAmt:        commontypes.DYM.MulRaw(10),
			initialBuyAmt:  math.ZeroInt(),
			planAllocation: commontypes.DYM.MulRaw(100),
			preGraduated:   true,
		},
		{
			name:           "pre-graduation buy - don't hit graduation",
			pumpAmt:        commontypes.DYM.MulRaw(10),
			initialBuyAmt:  commontypes.DYM.MulRaw(30),
			planAllocation: commontypes.DYM.MulRaw(100),
			preGraduated:   true,
		},
		{
			name:           "pre-graduation buy - triggers graduation",
			pumpAmt:        commontypes.DYM.MulRaw(50), // large pumpAmt to trigger graduation
			initialBuyAmt:  commontypes.DYM.MulRaw(70),
			planAllocation: commontypes.DYM.MulRaw(100),
			preGraduated:   true,
			graduated:      true,
		},
		{
			name:           "graduated buy - AMM swap",
			pumpAmt:        commontypes.DYM.MulRaw(10),
			initialBuyAmt:  commontypes.DYM.MulRaw(100), // buy the whole IRO to graduate it
			planAllocation: commontypes.DYM.MulRaw(100),
			graduated:      true,
		},
		{
			name:           "settled buy - AMM swap to rollapp token",
			pumpAmt:        commontypes.DYM.MulRaw(10),
			initialBuyAmt:  commontypes.DYM.MulRaw(100), // buy the whole IRO to graduate it
			planAllocation: commontypes.DYM.MulRaw(100),
			settled:        true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rollappID := s.CreateDefaultRollapp()
			planID := s.CreatePlan(rollappID, tc.planAllocation, false)
			if !tc.initialBuyAmt.IsZero() {
				s.buyIRO(planID, tc.initialBuyAmt)
			}
			if tc.settled {
				reserved := s.App.IROKeeper.MustGetPlan(s.Ctx, planID).SoldAmt
				s.SettleIRO(rollappID, reserved)
			}

			initialStreamerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))

			// Reset event manager
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())

			tokenOut, err := s.App.StreamerKeeper.ExecutePump(
				s.Ctx,
				tc.pumpAmt,
				sdk.DefaultBondDenom,
				rollappID,
			)
			s.Require().NoError(err)

			actualStreamerBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
			expectedStreamerBalance := initialStreamerBalance.Add(tokenOut).Sub(sdk.NewCoin(sdk.DefaultBondDenom, tc.pumpAmt))
			s.Require().Equal(expectedStreamerBalance, actualStreamerBalance, "expected %s, got %s", expectedStreamerBalance, actualStreamerBalance)

			plan := s.App.IROKeeper.MustGetPlan(s.Ctx, planID)

			switch {
			case tc.preGraduated && tc.graduated:
				s.Require().True(plan.IsGraduated())
				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(irotypes.EventBuy)), 1)
				s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtTokenSwapped, 1)
				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(irotypes.EventGraduation)), 1)

			case tc.preGraduated:
				s.Require().True(plan.PreGraduation())
				s.AssertEventEmitted(s.Ctx, proto.MessageName(new(irotypes.EventBuy)), 1)
				s.AssertEventNotEmitted(s.Ctx, gammtypes.TypeEvtTokenSwapped)
				s.AssertEventNotEmitted(s.Ctx, proto.MessageName(new(irotypes.EventGraduation)))

			case tc.graduated:
				s.Require().True(plan.IsGraduated())
				s.AssertEventNotEmitted(s.Ctx, proto.MessageName(new(irotypes.EventBuy)))
				s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtTokenSwapped, 1)
				s.AssertEventNotEmitted(s.Ctx, proto.MessageName(new(irotypes.EventGraduation)))

			case tc.settled:
				s.Require().True(plan.IsSettled())
				s.AssertEventNotEmitted(s.Ctx, proto.MessageName(new(irotypes.EventBuy)))
				s.AssertEventEmitted(s.Ctx, gammtypes.TypeEvtTokenSwapped, 1)
				s.AssertEventNotEmitted(s.Ctx, proto.MessageName(new(irotypes.EventGraduation)))
			}
		})
	}
}

// buyAmt is in IRO denom. If buyAmt > plan.MaxAmountToSell, then buy the entire plan and graduate.
func (s *KeeperTestSuite) buyIRO(planID string, buyAmt math.Int) {
	plan := s.App.IROKeeper.MustGetPlan(s.Ctx, planID)

	buyer := apptesting.CreateRandomAccounts(1)[0]
	s.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, commontypes.DYM.MulRaw(1000))))
	tokenIn := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(buyAmt))
	plusTakerFeeAmt, _, err := s.App.IROKeeper.ApplyTakerFee(tokenIn, s.App.IROKeeper.GetParams(s.Ctx).TakerFee, true)
	s.Require().NoError(err)
	s.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin(plan.LiquidityDenom, plusTakerFeeAmt)))

	if buyAmt.GT(plan.MaxAmountToSell.Sub(plan.SoldAmt)) {
		buyAmt = plan.MaxAmountToSell.Sub(plan.SoldAmt)
	}

	_, err = s.App.IROKeeper.Buy(s.Ctx, planID, buyer, buyAmt, plusTakerFeeAmt)
	s.Require().NoError(err)
}
