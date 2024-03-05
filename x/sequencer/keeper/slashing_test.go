package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) assertSlashed(seqAddr string) {
	seq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)
	suite.Assert().True(seq.Jailed)
	suite.Assert().Equal(types.Unbonded, seq.Status)
	suite.Assert().Equal(sdk.Coins(nil), seq.Tokens)

	sequencers := suite.App.SequencerKeeper.GetMatureUnbondingSequencers(suite.Ctx, suite.Ctx.BlockTime())
	for _, s := range sequencers {
		suite.Assert().NotEqual(s.SequencerAddress, seqAddr)
	}
}

func (suite *SequencerTestSuite) TestSlashingUnknownSequencer() {
	suite.SetupTest()

	_ = suite.CreateDefaultRollapp()
	keeper := suite.App.SequencerKeeper

	err := keeper.Slashing(suite.Ctx, "unknown_sequencer")
	suite.Assert().ErrorIs(err, types.ErrUnknownSequencer)
}

func (suite *SequencerTestSuite) TestSlashingUnbondedSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	res, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	unbondTime := res.CompletionTime
	keeper.UnbondAllMatureSequencers(suite.Ctx, unbondTime.Add(1*time.Second))

	seq, found := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)

	suite.Assert().Equal(seq.SequencerAddress, seqAddr)
	suite.Assert().Equal(seq.Status, types.Unbonded)
	err = keeper.Slashing(suite.Ctx, seqAddr)
	suite.Assert().ErrorIs(err, types.ErrInvalidSequencerStatus)
}

func (suite *SequencerTestSuite) TestSlashingUnbondingSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Assert().Equal(seq.Status, types.Unbonding)
	err = keeper.Slashing(suite.Ctx, seqAddr)
	suite.Assert().NoError(err)

	suite.assertSlashed(seqAddr)
}

func (suite *SequencerTestSuite) TestSlashingPropserSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seqAddr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Assert().Equal(seq.Status, types.Bonded)
	suite.Assert().True(seq.Proposer)

	seq2, ok := keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Assert().Equal(seq2.Status, types.Bonded)
	suite.Assert().False(seq2.Proposer)

	err := keeper.Slashing(suite.Ctx, seqAddr)
	suite.Assert().NoError(err)

	suite.assertSlashed(seqAddr)

	seq2, ok = keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Assert().Equal(seq2.Status, types.Bonded)
	suite.Assert().True(seq2.Proposer)
}
