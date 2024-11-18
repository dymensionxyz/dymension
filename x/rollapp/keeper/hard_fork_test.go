package keeper_test

import (
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// TestHardFork - Test the HardFork function
// - deleted states
// - pending queue is cleared up to the fraud height
// - revision number incremented
func (suite *RollappTestSuite) TestHardFork() {
	numOfSequencers := uint64(3) - 1 // already created one with rollapp
	numOfStates := uint64(100)
	numOfFinalizedStates := uint64(10)
	numOfBlocks := uint64(10)

	testCases := []struct {
		name            string
		statesCommitted uint64
		statesFinalized uint64
		fraudHeight     uint64
		expectError     bool
	}{
		// happy flows (fraud at different heights, states contains blocks 1-10, 11-20, 21-30, ...)
		{"Fraud at start of batch", numOfStates, numOfFinalizedStates, 101, false},
		{"Fraud in middle of batch", numOfStates, numOfFinalizedStates, 107, false},
		{"Fraud at end of batch", numOfStates, numOfFinalizedStates, 200, false},
		{"Fraud at future height", 10, 1, 300, false},

		// error flows
		{"first batch not committed yet", 0, 0, 10, true},
		{"first block of the first batch", 1, 0, 1, true},
		{"height already finalized", numOfStates, numOfFinalizedStates, 20, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset the state for the next test case
			suite.SetupTest()
			suite.App.RollappKeeper.SetHooks(nil) // disable hooks

			initialHeight := uint64(1)
			suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight))

			// unrelated rollapp just to validate it's unaffected
			rollapp2, proposer2 := suite.CreateDefaultRollappAndProposer()
			var (
				err        error
				lastHeight uint64 = 1
			)
			for i := uint64(0); i < numOfStates; i++ {
				suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight + i))
				lastHeight, err = suite.PostStateUpdate(suite.Ctx, rollapp2, proposer2, lastHeight, numOfBlocks)
				suite.Require().NoError(err)
			}

			// create rollapp and sequencers before fraud evidence
			rollappId, proposer := suite.CreateDefaultRollappAndProposer()
			for i := uint64(0); i < numOfSequencers-1; i++ {
				_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
			}

			// send state updates
			lastHeight = 1
			for i := uint64(0); i < tc.statesCommitted; i++ {
				suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialHeight + i))
				lastHeight, err = suite.PostStateUpdate(suite.Ctx, rollappId, proposer, lastHeight, numOfBlocks)
				suite.Require().NoError(err)
			}

			// Assert initial stats (revision 0, states pending)
			suite.assertNotForked(rollappId)
			queue, err := suite.App.RollappKeeper.GetFinalizationQueueByRollapp(suite.Ctx, rollappId)
			suite.Require().NoError(err)
			suite.Require().Len(queue, int(tc.statesCommitted))

			// finalize some of the states
			suite.App.RollappKeeper.FinalizeRollappStates(suite.Ctx.WithBlockHeight(int64(initialHeight + tc.statesFinalized)))

			err = suite.App.RollappKeeper.HardFork(suite.Ctx, rollappId, tc.fraudHeight)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.assertFraudHandled(rollappId, tc.fraudHeight)
			}
		})
	}
}

// Fail - Invalid rollapp
func (suite *RollappTestSuite) TestHardFork_InvalidRollapp() {
	ctx := &suite.Ctx
	keeper := suite.App.RollappKeeper

	rollapp, proposer := suite.CreateDefaultRollappAndProposer()
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)

	err = keeper.HardFork(*ctx, "invalidRollapp", 2)
	suite.Require().Error(err)
}

// Fail - Disputing already finalized state
func (suite *RollappTestSuite) TestHardFork_AlreadyFinalized() {
	ctx := &suite.Ctx
	keeper := suite.App.RollappKeeper

	rollapp, proposer := suite.CreateDefaultRollappAndProposer()
	_, err := suite.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	suite.Require().Nil(err)

	// finalize state
	suite.Ctx = suite.Ctx.WithBlockHeight(ctx.BlockHeight() + int64(keeper.DisputePeriodInBlocks(*ctx)))
	suite.App.RollappKeeper.FinalizeRollappStates(suite.Ctx)
	stateInfo, err := suite.App.RollappKeeper.FindStateInfoByHeight(suite.Ctx, rollapp, 1)
	suite.Require().Nil(err)
	suite.Require().Equal(common.Status_FINALIZED, stateInfo.Status)

	err = keeper.HardFork(*ctx, rollapp, 2)
	suite.Require().NotNil(err)
}

/* ---------------------------------- utils --------------------------------- */
func (suite *RollappTestSuite) assertFraudHandled(rollappId string, height uint64) {
	rollapp, found := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappId)
	suite.Require().True(found)
	suite.Require().Equal(uint64(1), rollapp.LatestRevision().Number)

	// check states were deleted
	// the last state should have height less than the fraud height
	lastestStateInfo, ok := suite.App.RollappKeeper.GetLatestStateInfo(suite.Ctx, rollappId)
	if ok {
		suite.Require().Less(lastestStateInfo.GetLatestHeight(), height)
	}

	// check sequencers heights
	sequencers, err := suite.App.RollappKeeper.AllSequencerHeightPairs(suite.Ctx)
	suite.Require().NoError(err)

	ok = false
	for _, seq := range sequencers {
		if seq.Sequencer == lastestStateInfo.Sequencer {
			suite.Require().Less(seq.Height, height)
			ok = true
		}
	}
	suite.Require().True(ok)

	// check queue
	queue, err := suite.App.RollappKeeper.GetFinalizationQueueByRollapp(suite.Ctx, rollappId)
	suite.Require().NoError(err)
	suite.Require().Greater(len(queue), 0)
	for _, q := range queue {
		for _, stateInfoIndex := range q.FinalizationQueue {
			if stateInfoIndex.RollappId == rollappId {
				suite.Require().LessOrEqual(stateInfoIndex.Index, lastestStateInfo.StateInfoIndex.Index)
			}
		}
	}
}
