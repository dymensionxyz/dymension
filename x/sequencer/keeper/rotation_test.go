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
			currBond := bond
			for i := 0; i < len(seqAddrs); i++ {
				currBond = currBond.AddAmount(bond.Amount)
				seqAddrs[i] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, currBond)
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

// tests that the proposer and nextProposer are filtered out from the list of expected next proposers
func (suite *SequencerTestSuite) TestExpectedNextProposerFiltering() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	_ = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond) // proposer
	seqAddrs := make([]string, 4)
	currBond := bond
	for i := 0; i < len(seqAddrs); i++ {
		currBond = currBond.AddAmount(bond.Amount)
		seqAddrs[i] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, currBond)
	}

	next := suite.App.SequencerKeeper.ExpectedNextProposer(suite.Ctx, rollappId)
	suite.Equal(seqAddrs[len(seqAddrs)-1], next.SequencerAddress)

	// set the expected next proposer as the next proposer
	suite.App.SequencerKeeper.SetNextProposer(suite.Ctx, rollappId, seqAddrs[len(seqAddrs)-1])

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
	res, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// mature notice period
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, res.GetNoticePeriodCompletionTime().Add(10*time.Second))
	p, ok := suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)

	// simulate lastBlock received
	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId)

	// assert addr2 is now proposer
	p, ok = suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr2, p.SequencerAddress)
	// assert addr1 is unbonding
	u, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonding, u.Status)
	// assert nextProposer is nil
	_, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)
}

func (suite *SequencerTestSuite) TestRotateProposerNoNextProposer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: addr1}
	res, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// mature notice period
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, res.GetNoticePeriodCompletionTime().Add(10*time.Second))
	// simulate lastBlock received
	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId)

	_, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)

	_, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)
}

// Both the proposer and nextProposer tries to unbond
func (suite *SequencerTestSuite) TestStartRotationTwice() {
	suite.SetupTest()
	suite.Ctx = suite.Ctx.WithBlockHeight(10)

	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	addr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	// unbond proposer
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	p, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr1, p.SequencerAddress)
	suite.Equal(suite.Ctx.BlockHeight(), p.UnbondRequestHeight)

	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, p.UnbondTime.Add(10*time.Second))
	suite.Require().True(suite.App.SequencerKeeper.IsRotating(suite.Ctx, rollappId))

	n, ok := suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr2, n.SequencerAddress)

	// unbond nextProposer before rotation completes
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check nextProposer is still the nextProposer and notice period started
	n, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(addr2, n.SequencerAddress)
	suite.Require().True(n.IsNoticePeriodInProgress())

	// rotation completes before notice period ends for addr2 (the nextProposer)
	// FIXME: what if notice period ends for addr2 before rotation completes?
	suite.App.SequencerKeeper.RotateProposer(suite.Ctx, rollappId) // simulate lastBlock received

	// validate addr2 is now proposer and still with notice period
	p, _ = suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Equal(addr2, p.SequencerAddress)
	suite.Require().True(p.IsNoticePeriodInProgress())

	// validate nextProposer is unset after rotation completes
	n, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)

	// mature notice period for addr2
	suite.App.SequencerKeeper.MatureSequencersWithNoticePeriod(suite.Ctx, p.UnbondTime.Add(10*time.Second))
	// validate nextProposer is set
	n, ok = suite.App.SequencerKeeper.GetNextProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Require().Empty(n.SequencerAddress)
}
