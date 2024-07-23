package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestInvariants() {
	suite.SetupTest()
	initialheight := uint64(10)
	initialTime := time.Now()

	numOfRollapps := 5
	numOfSequencers := 5

	var (
		rollappToTest string
		timeToMature  time.Time
	)

	// create rollapps and sequencers
	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()

		// create sequencers
		seqAddr := make([]string, numOfSequencers)
		for j := 0; j < numOfSequencers; j++ {
			seqAddr[j] = suite.CreateDefaultSequencer(suite.Ctx, rollapp)
		}

		// unbonding some sequencers
		for j := uint64(0); j < uint64(numOfSequencers-1); j++ {
			suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + j)).WithBlockTime(initialTime.Add(time.Duration(j) * time.Second))
			res, err := suite.msgServer.Unbond(suite.Ctx, &types.MsgUnbond{Creator: seqAddr[j]})
			suite.Require().NoError(err)
			if i == 1 && j == 1 {
				rollappToTest = rollapp
				timeToMature = *res.GetUnbondingCompletionTime()
			}
		}
	}

	rollappid := rollappToTest
	seqUnbonding := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Unbonding)
	suite.Require().True(len(seqUnbonding) > 0)

	// unbond some unbonding sequencers
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, timeToMature)

	// Test the test: make sure all status have entries
	seqBonded := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Bonded)
	seqUnbonding = suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Unbonding)
	seqUnbonded := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Unbonded)

	if len(seqBonded) == 0 || len(seqUnbonding) == 0 || len(seqUnbonded) == 0 {
		suite.T().Fatal("Test setup failed")
	}
	// additional rollapp with no sequencers
	_ = suite.CreateDefaultRollapp()

	msg, bool := keeper.AllInvariants(suite.App.SequencerKeeper)(suite.Ctx)
	suite.Require().False(bool, msg)
}
