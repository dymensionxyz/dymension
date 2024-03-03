package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingMultiple() {
	suite.SetupTest()
	suite.Ctx = suite.Ctx.WithBlockHeight(10)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	rollappId2 := suite.CreateDefaultRollapp()

	numOfSequencers := 5
	numOfSequencers2 := 3
	unbodingSeq := 2

	seqAddr1 := make([]string, numOfSequencers)
	seqAddr2 := make([]string, numOfSequencers2)

	// create 5 sequencers for rollapp1
	for i := 0; i < numOfSequencers; i++ {
		seqAddr1[i] = suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	}

	// create 3 sequencers for rollapp2
	for i := 0; i < numOfSequencers2; i++ {
		seqAddr2[i] = suite.CreateDefaultSequencer(suite.Ctx, rollappId2)
	}

	// start unbonding for 2 sequencers in each rollapp
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	now := time.Now()
	unbondTime := now.Add(keeper.GetParams(suite.Ctx).UnbondingTime)
	suite.Ctx = suite.Ctx.WithBlockTime(now)
	for i := 0; i < unbodingSeq; i++ {
		unbondMsg := types.MsgUnbond{Creator: seqAddr1[i]}
		_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
		suite.Require().NoError(err)

		unbondMsg = types.MsgUnbond{Creator: seqAddr2[i]}
		_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
		suite.Require().NoError(err)
	}

	// before unbonding time reached
	sequencers := keeper.GetMatureUnbondingSequencers(suite.Ctx, now)
	suite.Require().Len(sequencers, 0)

	sequencers = keeper.GetMatureUnbondingSequencers(suite.Ctx, unbondTime.Add(-1*time.Second))
	suite.Require().Len(sequencers, 0)

	// past unbonding time
	sequencers = keeper.GetMatureUnbondingSequencers(suite.Ctx, unbondTime.Add(1*time.Second))
	suite.Require().Len(sequencers, 4)
}

func (suite *SequencerTestSuite) TestTokensRefundOnUnbond() {
	suite.SetupTest()
	denom := bond.Denom
	blockheight := 20
	var err error

	rollappId := suite.CreateDefaultRollapp()
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	sequencer1, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(sequencer1.Status == types.Bonded)
	suite.Require().True(sequencer1.Proposer)

	suite.Require().False(sequencer1.Tokens.IsZero())

	addr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	sequencer2, _ := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(sequencer2.Status == types.Bonded)
	suite.Require().False(sequencer2.Proposer)

	suite.Require().False(sequencer2.Tokens.IsZero())

	suite.Ctx = suite.Ctx.WithBlockHeight(int64(blockheight))
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	//start the 1st unbond
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)
	sequencer1, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(sequencer1.Status == types.Unbonding)
	suite.Require().Equal(sequencer1.UnbondingHeight, int64(blockheight))
	suite.Require().False(sequencer1.Tokens.IsZero())

	//start the 2nd unbond later
	suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeight() + 1)
	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(5 * time.Minute))
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)
	sequencer2, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(sequencer2.Status == types.Unbonding)
	suite.Require().False(sequencer2.Tokens.IsZero())

	/* -------------------------- check the unbond phase ------------------------- */
	balanceBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, sdk.MustAccAddressFromBech32(addr1), denom)
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, sequencer1.UnbondTime.Add(1*time.Second))
	balanceAfter := suite.App.BankKeeper.GetBalance(suite.Ctx, sdk.MustAccAddressFromBech32(addr1), denom)

	//Check stake refunded
	sequencer1, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Equal(types.Unbonded, sequencer1.Status)
	suite.True(sequencer1.Tokens.IsZero())
	suite.True(balanceBefore.Add(bond).IsEqual(balanceAfter), "expected %s, got %s", balanceBefore.Add(bond), balanceAfter)

	//check the 2nd unbond still not happened
	sequencer2, _ = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Equal(types.Unbonding, sequencer2.Status)
	suite.False(sequencer2.Tokens.IsZero())
}
