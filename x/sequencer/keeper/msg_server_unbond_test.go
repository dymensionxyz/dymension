package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestUnbondingNonProposer() {
	rollappId, pk := suite.CreateDefaultRollapp()
	proposerAddr := suite.CreateSequencer(suite.Ctx, rollappId, pk)

	bondedAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	suite.Require().NotEqual(proposerAddr, bondedAddr)

	proposer, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(proposerAddr, proposer.Address)

	/* ------------------------- unbond non proposer sequencer ------------------------ */
	bondedSeq, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, bondedAddr)
	suite.Require().True(found)
	suite.Equal(types.Bonded, bondedSeq.Status)

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

	// check proposer not changed
	proposer, ok = suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(proposerAddr, proposer.Address)

	// try to unbond again. already unbonded, we expect error
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)
}

func (suite *SequencerTestSuite) TestUnbondingProposer() {
	suite.Ctx = suite.Ctx.WithBlockHeight(10)

	rollappId, proposerAddr := suite.CreateDefaultRollappAndProposer()
	_ = suite.CreateSequencer(suite.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	/* ----------------------------- unbond proposer ---------------------------- */
	unbondMsg := types.MsgUnbond{Creator: proposerAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	// check proposer still bonded and notice period started
	p, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(proposerAddr, p.Address)
	suite.Equal(suite.Ctx.BlockHeight(), p.UnbondRequestHeight)

	// unbonding again, we expect error as sequencer is in notice period
	_, err = suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().Error(err)

	// next proposer should not be set yet
	_, ok = suite.App.SequencerKeeper.GetNextProposerAddr(suite.Ctx, rollappId)
	suite.Require().False(ok)

	// check notice period queue
	m := suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.NoticePeriodTime.Add(-1*time.Second))
	suite.Require().Len(m, 0)
	m = suite.App.SequencerKeeper.GetMatureNoticePeriodSequencers(suite.Ctx, p.NoticePeriodTime.Add(1*time.Second))
	suite.Require().Len(m, 1)
}
