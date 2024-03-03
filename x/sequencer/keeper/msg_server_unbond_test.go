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
	addr3 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	/* ----------------------------- unbond proposer ---------------------------- */
	sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer.Status)
	suite.True(sequencer.Proposer)

	sequencer2, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer2.Status)
	suite.False(sequencer2.Proposer)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check proposer rotation
	sequencer2, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Equal(types.Bonded, sequencer2.Status)
	suite.True(sequencer2.Proposer)

	// check sequencer operating status
	sequencer, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonding, sequencer.Status)
	suite.False(sequencer.Proposer)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer.UnbondTime.Add(10*time.Second))

	sequencer, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonded, sequencer.Status)

	/* ------------------------- unbond non proposer sequencer ------------------------ */
	sequencer3, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Bonded, sequencer3.Status)
	suite.False(sequencer3.Proposer)

	unbondMsg = types.MsgUnbond{Creator: addr3}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check sequencer operating status
	sequencer3, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Unbonding, sequencer3.Status)

	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer3.UnbondTime.Add(10*time.Second))

	sequencer3, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr3)
	suite.Require().True(found)
	suite.Equal(types.Unbonded, sequencer3.Status)
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
