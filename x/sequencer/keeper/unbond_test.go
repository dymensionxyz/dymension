package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingMultiple() {
	suite.SetupTest()
	suite.ctx = suite.ctx.WithBlockHeight(10)
	suite.ctx = suite.ctx.WithBlockTime(time.Now())

	keeper := suite.app.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	rollappId2 := suite.CreateDefaultRollapp()

	numOfSequencers := 5
	numOfSequencers2 := 3
	unbodingSeq := 2

	seqAddr1 := make([]string, numOfSequencers)
	seqAddr2 := make([]string, numOfSequencers2)

	// create 5 sequencers for rollapp1
	for i := 0; i < numOfSequencers; i++ {
		seqAddr1[i] = suite.CreateDefaultSequencer(suite.ctx, rollappId)
	}

	// create 3 sequencers for rollapp2
	for i := 0; i < numOfSequencers2; i++ {
		seqAddr2[i] = suite.CreateDefaultSequencer(suite.ctx, rollappId2)
	}

	// start unbonding for 2 sequencers in each rollapp
	suite.ctx = suite.ctx.WithBlockHeight(20)
	now := time.Now()
	unbondTime := now.Add(keeper.GetParams(suite.ctx).UnbondingTime)
	suite.ctx = suite.ctx.WithBlockTime(now)
	for i := 0; i < unbodingSeq; i++ {
		unbondMsg := types.MsgUnbond{Creator: seqAddr1[i]}
		_, err := suite.msgServer.Unbond(suite.ctx, &unbondMsg)
		suite.Require().NoError(err)

		unbondMsg = types.MsgUnbond{Creator: seqAddr2[i]}
		_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
		suite.Require().NoError(err)
	}

	// before unbonding time reached
	sequencers := keeper.GetMatureUnbondingSequencers(suite.ctx, now)
	suite.Require().Len(sequencers, 0)

	sequencers = keeper.GetMatureUnbondingSequencers(suite.ctx, unbondTime.Add(-1*time.Second))
	suite.Require().Len(sequencers, 0)

	// past unbonding time
	sequencers = keeper.GetMatureUnbondingSequencers(suite.ctx, unbondTime.Add(1*time.Second))
	suite.Require().Len(sequencers, 4)
}

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
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(5 * time.Minute))
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.ctx, &unbondMsg)
	suite.Require().NoError(err)
	sequencer2, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(sequencer2.Status == types.Unbonding)
	suite.Require().False(sequencer2.Tokens.IsZero())

	/* -------------------------- check the unbond phase ------------------------- */
	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(addr1), denom)
	suite.app.SequencerKeeper.UnbondAllMatureSequencers(suite.ctx, sequencer1.UnbondTime.Add(1*time.Second))
	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(addr1), denom)

	//Check stake refunded
	sequencer1, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Equal(types.Unbonded, sequencer1.Status)
	suite.True(sequencer1.Tokens.IsZero())
	suite.True(balanceBefore.Add(bond).IsEqual(balanceAfter), "expected %s, got %s", balanceBefore.Add(bond), balanceAfter)

	//check the 2nd unbond still not happened
	sequencer2, _ = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Equal(types.Unbonding, sequencer2.Status)
	suite.False(sequencer2.Tokens.IsZero())
}
