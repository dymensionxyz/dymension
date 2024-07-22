package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestExpectedNextProposer() {
	type testCase struct {
		name                    string
		numSeqAddrs             int
		expectEmptyNextProposer bool
	}

	testCases := []testCase{
		{"No additional sequencers", 0, true},
		{"few", 4, false},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			rollappId := suite.CreateDefaultRollapp()

			_ = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond) // proposer
			seqAddrs := make([]string, tc.numSeqAddrs)
			for i := 0; i < len(seqAddrs); i++ {
				seqAddrs[i] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(bond.Amount))
			}
			next := suite.App.SequencerKeeper.ExpectedNextProposer(suite.Ctx, rollappId)
			if tc.expectEmptyNextProposer {
				suite.Require().Empty(next.SequencerAddress)
				return
			}

			expectedNextProposer := seqAddrs[len(seqAddrs)-1]
			suite.Equal(expectedNextProposer, next.SequencerAddress)
		})
	}
}

// tests that current next proposer is filtered out from the list of expected next proposers
func (suite *SequencerTestSuite) TestExpectedNextProposerFiltering() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	_ = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond) // proposer
	seqAddrs := make([]string, 10)
	for i := 0; i < len(seqAddrs); i++ {
		seqAddrs[i] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(bond.Amount))
	}

	next := suite.App.SequencerKeeper.ExpectedNextProposer(suite.Ctx, rollappId)
	suite.Equal(seqAddrs[len(seqAddrs)-1], next.SequencerAddress)

	// set the expected next proposer as the next proposer
	suite.App.SequencerKeeper.SetNextProposer(suite.Ctx, rollappId, types.Sequencer{SequencerAddress: seqAddrs[len(seqAddrs)-1]})

	// check that the next proposer is filtered out
	next = suite.App.SequencerKeeper.ExpectedNextProposer(suite.Ctx, rollappId)
	suite.Equal(seqAddrs[len(seqAddrs)-2], next.SequencerAddress)

}

// TestStartRotation tests the StartRotation function which is called when a sequencer has finished its notice period
func (suite *SequencerTestSuite) TestStartRotation() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check proposer still bonded and notice period started
	p, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr1, p.SequencerAddress)
	suite.Equal(suite.Ctx.BlockHeight(), p.UnbondRequestHeight)

	m := suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.UnbondTime.Add(-10*time.Second))
	suite.Require().Len(m, 0)
	m = suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.UnbondTime.Add(10*time.Second))
	suite.Require().Len(m, 1)
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, p.UnbondTime.Add(10*time.Second))

	// validate nextProposer is set
	n, ok := suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().NotEmpty(n.SequencerAddress)

	// validate proposer not changed
	p, _ = suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Equal(addr1, p.SequencerAddress)

}

func (suite *SequencerTestSuite) TestRotateProposer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	addr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// mature notice period
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, suite.Ctx.BlockTime().Add(10*time.Second))
	// simulate lastBlock received
	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId)

	// assert addr2 is now proposer
	p, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr2, p.SequencerAddress)
	// assert addr1 is unbonding
	u, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonding, u.Status)
	// assert nextProposer is nil
	_, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)
}

func (suite *SequencerTestSuite) TestRotateProposerEmptyNextProposer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// mature notice period
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, suite.Ctx.BlockTime().Add(10*time.Second))
	// simulate lastBlock received
	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId)

	next, ok := suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().Empty(next.SequencerAddress)
}

// TODO: test nextSequencer also unbonds
