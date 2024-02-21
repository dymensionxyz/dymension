package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingStatusChange() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.ctx, rollappId)
	addr2 := suite.CreateDefaultSequencer(suite.ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	sequencer, found := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Proposer)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer, found = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Unbonding)

	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondTime.Add(10*time.Second))

	sequencer, found = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Unbonded)

	/* ------------------------- unbond bonded sequencer ------------------------ */
	sequencer2, found := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Bonded)

	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer2, found = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Unbonding)

	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer2.UnbondTime.Add(10*time.Second))

	sequencer2, found = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Unbonded)
}

func (suite *SequencerTestSuite) TestUnbondingNotBondedSequencer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.ctx, rollappId)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)

	//already unbonding, we expect error
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().Error(err)

	sequencer, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondTime.Add(10*time.Second))

	//already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().Error(err)

}
