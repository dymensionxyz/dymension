package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingNonProposer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	bondedAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ------------------------- unbond non proposer sequencer ------------------------ */
	bondedSeq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, bondedAddr)
	suite.Require().True(found)
	suite.Equal(types.Bonded, bondedSeq.Status)
	suite.False(bondedSeq.Proposer)

	unbondMsg := types.MsgUnbond{Creator: bondedAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	bondedSeq, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, bondedAddr)
	suite.Require().True(found)
	suite.Equal(types.Unbonding, bondedSeq.Status)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, bondedSeq.UnbondTime.Add(10*time.Second))
	bondedSeq, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, bondedAddr)
	suite.Require().True(found)
	suite.Equal(types.Unbonded, bondedSeq.Status)
}

func (suite *SequencerTestSuite) TestUnbondingProposer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

	proposerAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: proposerAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check proposer still bonded and notice period started
	p := suite.App.SequencerKeeper.GetRollappProposer(suite.Ctx, rollappId)
	suite.Require().NotNil(p)
	suite.Equal(proposerAddr, p.SequencerAddress)
	suite.Equal(suite.Ctx.BlockHeight(), p.UnbondRequestHeight)
	n := suite.App.SequencerKeeper.GetRollappNextProposer(suite.Ctx, rollappId)
	suite.Nil(n)

	m := suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.UnbondTime.Add(-10*time.Second))
	suite.Require().Len(m, 0)
	m = suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.UnbondTime.Add(10*time.Second))
	suite.Require().Len(m, 1)
}

func (suite *SequencerTestSuite) TestUnbondingNotBondedSequencer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// already unbonding, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)

	sequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	// already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)
}
