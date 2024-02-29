package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingStatusChange() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	addr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Proposer)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Unbonding)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	sequencer, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.Equal(sequencer.Status, types.Unbonded)

	/* ------------------------- unbond bonded sequencer ------------------------ */
	sequencer2, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Bonded)

	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer2, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Unbonding)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer2.UnbondTime.Add(10*time.Second))

	sequencer2, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.Equal(sequencer2.Status, types.Unbonded)
}

func (suite *SequencerTestSuite) TestUnbondingNotBondedSequencer() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	//already unbonding, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)

	sequencer, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	//already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)

}
