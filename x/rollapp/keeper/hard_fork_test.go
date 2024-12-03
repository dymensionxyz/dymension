package keeper_test

import (
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

// TestHardFork - Test the HardFork function
// - deleted states
// - pending queue is cleared up to the fraud height
// - revision number incremented
func (s *RollappTestSuite) TestHardFork() {
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
		s.Run(tc.name, func() {
			// Reset the state for the next test case
			s.SetupTest()
			s.k().SetHooks(nil) // disable hooks

			initialHeight := uint64(1)
			s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight))

			// unrelated rollapp just to validate it's unaffected
			rollapp2, proposer2 := s.CreateDefaultRollappAndProposer()
			var (
				err        error
				lastHeight uint64 = 1
			)
			for i := uint64(0); i < numOfStates; i++ {
				s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight + i))
				lastHeight, err = s.PostStateUpdate(s.Ctx, rollapp2, proposer2, lastHeight, numOfBlocks)
				s.Require().NoError(err)
			}

			// create rollapp and sequencers before fraud evidence
			rollappId, proposer := s.CreateDefaultRollappAndProposer()
			for i := uint64(0); i < numOfSequencers-1; i++ {
				_ = s.CreateDefaultSequencer(s.Ctx, rollappId)
			}

			ra := s.k().MustGetRollapp(s.Ctx, rollappId)
			ra.GenesisState.TransferProofHeight = 1
			s.k().SetRollapp(s.Ctx, ra)

			// send state updates
			lastHeight = 1
			for i := uint64(0); i < tc.statesCommitted; i++ {
				s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight + i))
				lastHeight, err = s.PostStateUpdate(s.Ctx, rollappId, proposer, lastHeight, numOfBlocks)
				s.Require().NoError(err)
			}

			// Assert initial stats (revision 0, states pending)
			s.assertNotForked(rollappId)
			queue, err := s.k().GetFinalizationQueueByRollapp(s.Ctx, rollappId)
			s.Require().NoError(err)
			s.Require().Len(queue, int(tc.statesCommitted))

			// finalize some of the states
			s.k().FinalizeRollappStates(s.Ctx.WithBlockHeight(int64(initialHeight + tc.statesFinalized)))

			err = s.k().HardFork(s.Ctx, rollappId, tc.fraudHeight-1)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.assertFraudHandled(rollappId, tc.fraudHeight)
				s.checkLiveness(rollappId, true, false)
			}
		})
	}
}

// Fail - Invalid rollapp
func (s *RollappTestSuite) TestHardFork_InvalidRollapp() {
	ctx := &s.Ctx

	rollapp, proposer := s.CreateDefaultRollappAndProposer()
	_, err := s.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	s.Require().Nil(err)

	err = s.k().HardFork(*ctx, "invalidRollapp", 1)
	s.Require().Error(err)
}

func (s *RollappTestSuite) TestAfterSetRealProposer() {
	ctx := &s.Ctx
	rollapp, proposer := s.CreateDefaultRollappAndProposer()
	_, err := s.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	s.Require().NoError(err)
	h := &keeper.SequencerHooks{
		Keeper: s.k(),
	}
	err = h.AfterSetRealProposer(*ctx, rollapp, s.App.SequencerKeeper.GetProposer(*ctx, rollapp))
	s.Require().NoError(err)
	s.checkLiveness(rollapp, true, true)
}

// Fail - Disputing already finalized state
func (s *RollappTestSuite) TestHardFork_AlreadyFinalized() {
	ctx := &s.Ctx

	rollapp, proposer := s.CreateDefaultRollappAndProposer()
	_, err := s.PostStateUpdate(*ctx, rollapp, proposer, 1, uint64(10))
	s.Require().Nil(err)

	// finalize state
	s.Ctx = s.Ctx.WithBlockHeight(ctx.BlockHeight() + int64(s.k().DisputePeriodInBlocks(*ctx)))
	s.k().FinalizeRollappStates(s.Ctx)
	stateInfo, err := s.k().FindStateInfoByHeight(s.Ctx, rollapp, 1)
	s.Require().Nil(err)
	s.Require().Equal(common.Status_FINALIZED, stateInfo.Status)

	err = s.k().HardFork(*ctx, rollapp, 1)
	s.Require().NotNil(err)
}

/* ---------------------------------- utils --------------------------------- */
func (s *RollappTestSuite) assertFraudHandled(rollappId string, height uint64) {
	rollapp, found := s.k().GetRollapp(s.Ctx, rollappId)
	s.Require().True(found)
	s.Require().Equal(uint64(1), rollapp.LatestRevision().Number)

	// check states were deleted
	// the last state should have height less than the fraud height
	lastestStateInfo, ok := s.k().GetLatestStateInfo(s.Ctx, rollappId)
	if ok {
		s.Require().Less(lastestStateInfo.GetLatestHeight(), height)
	}

	// check sequencers heights
	sequencers, err := s.k().AllSequencerHeightPairs(s.Ctx)
	s.Require().NoError(err)

	ok = false
	for _, seq := range sequencers {
		if seq.Sequencer == lastestStateInfo.Sequencer {
			s.Require().Less(seq.Height, height)
			ok = true
		}
	}
	s.Require().True(ok)

	// check queue
	queue, err := s.k().GetFinalizationQueueByRollapp(s.Ctx, rollappId)
	s.Require().NoError(err)
	s.Require().Greater(len(queue), 0)
	for _, q := range queue {
		for _, stateInfoIndex := range q.FinalizationQueue {
			if stateInfoIndex.RollappId == rollappId {
				s.Require().LessOrEqual(stateInfoIndex.Index, lastestStateInfo.StateInfoIndex.Index)
			}
		}
	}
}
