package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

//TODO: test multiple unbonds in parallal

func (suite *SequencerTestSuite) TestTokensRefundOnUnbond() {
	suite.SetupTest()
	denom := bond.Denom
	blockheight := 20
	var err error

	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.ctx, rollappId)
	sequencer1, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(sequencer1.Status == types.Proposer)
	suite.Require().False(sequencer1.Tokens.IsZero())

	addr2 := suite.CreateDefaultSequencer(suite.ctx, rollappId)
	sequencer2, _ := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(sequencer2.Status == types.Bonded)
	suite.Require().False(sequencer2.Tokens.IsZero())

	suite.ctx = suite.ctx.WithBlockHeight(int64(blockheight))
	suite.ctx = suite.ctx.WithBlockTime(time.Now())

	//start the 1st unbond
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)
	sequencer1, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(sequencer1.Status == types.Unbonding)
	suite.Require().Equal(sequencer1.UnbondingHeight, int64(blockheight))
	suite.Require().False(sequencer1.Tokens.IsZero())

	//start the 2nd unbond later
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(200 * time.Second))
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)
	sequencer2, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(sequencer2.Status == types.Unbonding)
	suite.Require().False(sequencer2.Tokens.IsZero())

	/* -------------------------- check the unbond phase ------------------------- */
	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(addr1), denom)
	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer1.UnbondingTime.Add(1*time.Second))
	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(addr1), denom)

	//Check stake refunded
	sequencer1, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Equal(sequencer1.Status, types.Unbonded)
	suite.True(sequencer1.Tokens.IsZero())
	suite.True(balanceBefore.Add(bond).IsEqual(balanceAfter), "expected %s, got %s", balanceBefore.Add(bond), balanceAfter)

	//check the 2nd unbond still not happened
	sequencer2, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Equal(sequencer2.Status, types.Unbonding)
	suite.False(sequencer2.Tokens.IsZero())
}
