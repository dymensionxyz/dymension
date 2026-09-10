package keeper_test

import (
	"cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// TODO: test slashing and Rewardee

func (s *RollappTestSuite) TestSubmitRollappFraud() {
	numOfBlocks := uint64(10)
	initialHeight := uint64(1)

	testCases := []struct {
		name            string
		msgRevision     uint64
		msgHeight       uint64
		expectedError   bool
		removeRevisions bool
	}{
		{
			name:          "first revision proposal",
			msgRevision:   0,
			msgHeight:     40,
			expectedError: false,
		},
		{
			name:          "future proposal",
			msgRevision:   2,
			msgHeight:     300,
			expectedError: false,
		},
		{
			name:          "wrong revision",
			msgRevision:   0,
			msgHeight:     80,
			expectedError: true,
		},
		{
			name:          "wrong future revision",
			msgRevision:   5,
			msgHeight:     300,
			expectedError: true,
		},
		{
			name:            "height without revision",
			msgRevision:     0,
			msgHeight:       40,
			expectedError:   true,
			removeRevisions: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // Reset state for each test case
			s.k().SetHooks(nil)

			s.Ctx = s.Ctx.WithBlockHeight(int64(initialHeight))
			rollappId, proposer := s.CreateDefaultRollappAndProposer()

			// set transferEnabled to true
			rollapp := s.k().MustGetRollapp(s.Ctx, rollappId)
			rollapp.GenesisState.TransferProofHeight = 1
			s.k().SetRollapp(s.Ctx, rollapp)

			var (
				lastHeight uint64 = 1
				err        error
			)

			// Create first batch of states (1-100)
			for i := int64(0); i < 10; i++ {
				s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + i)
				lastHeight, err = s.PostStateUpdate(s.Ctx, rollappId, proposer, lastHeight, numOfBlocks)
				s.Require().NoError(err)
			}

			// Force a fork at height 50
			err = s.k().HardFork(s.Ctx, rollappId, 49)
			s.Require().NoError(err)
			lastHeight = 50

			// Create more states after fork (51-150)
			for i := int64(10); i < 15; i++ {
				s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + i)
				lastHeight, err = s.PostStateUpdateWithRevision(s.Ctx, rollappId, proposer, lastHeight, numOfBlocks, 1)
				s.Require().NoError(err)
			}

			// Force another fork at height 120
			err = s.k().HardFork(s.Ctx, rollappId, 119)
			s.Require().NoError(err)

			// assert revision correctness
			rollapp = s.k().MustGetRollapp(s.Ctx, rollappId)
			for height, expectedRevision := range map[uint64]uint64{1: 0, 49: 0, 50: 1, 55: 1, 120: 2, 300: 2} {
				revision, found := rollapp.GetRevisionForHeight(height)
				s.Require().True(found)
				s.Require().EqualValues(expectedRevision, revision.Number)
			}

			if tc.removeRevisions {
				rollapp.Revisions = nil
				s.k().SetRollapp(s.Ctx, rollapp)
			}

			msg := &types.MsgRollappFraudProposal{
				Authority:              s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String(),
				RollappId:              rollappId,
				FraudHeight:            tc.msgHeight,
				FraudRevision:          tc.msgRevision,
				PunishSequencerAddress: "",
				Rewardee:               "",
			}

			_, err = s.k().SubmitRollappFraud(s.Ctx, msg)

			if !tc.expectedError {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				if tc.removeRevisions {
					s.Require().True(errors.IsOf(err, gerrc.ErrFailedPrecondition))
				}
			}
		})
	}
}

func (s *RollappTestSuite) TestSubmitRollappFraudPreservesValidationError() {
	msg := &types.MsgRollappFraudProposal{
		Authority: s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String(),
	}

	_, err := s.k().SubmitRollappFraud(s.Ctx, msg)

	s.Require().ErrorContains(err, "rollapp ID is invalid")
}
