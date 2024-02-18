package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

//TODO: check min bond

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

	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondingTime.Add(10*time.Second))

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

	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer2.UnbondingTime.Add(10*time.Second))

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
	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondingTime.Add(10*time.Second))

	//already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().Error(err)

}

// FIXME: test with acutal BOND
func (suite *SequencerTestSuite) TestTokensRefund() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.ctx, rollappId)
	addr2 := suite.CreateDefaultSequencer(suite.ctx, rollappId)

	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err := suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)

	//start the 2nd unbond later
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(200 * time.Second))
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)

	sequencer, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondingTime.Add(1*time.Second))

	//Check stake refunded
	sequencer, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Equal(sequencer.Status, types.Unbonded)
	suite.True(sequencer.Tokens.IsZero())
	suite.Equal(suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(addr1), sdk.DefaultBondDenom).Amount, sdk.ZeroInt())

	sequencer2, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Equal(sequencer2.Status, types.Unbonding)
	suite.False(sequencer2.Tokens.IsZero())
}
