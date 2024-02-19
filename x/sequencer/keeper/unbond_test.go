package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

//TODO: test multiple unbonds

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
	suite.Require().True(sequencer.Status == types.Unbonding)
	suite.Require().False(sequencer.Tokens.IsZero())

	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(addr1), sequencer.Tokens.Denom)

	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer.UnbondingTime.Add(1*time.Second))

	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.AccAddress(addr1), sequencer.Tokens.Denom)

	//Check stake refunded
	sequencer, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Equal(sequencer.Status, types.Unbonded)
	suite.True(sequencer.Tokens.IsZero())
	suite.Equal(balanceBefore.Add(sequencer.Tokens), balanceAfter)

	//check the 2nd unbond still not happened
	sequencer2, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Equal(sequencer2.Status, types.Unbonding)
	suite.False(sequencer2.Tokens.IsZero())
}
