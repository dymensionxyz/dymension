package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestSlashingUnknownSequencer() {
	suite.CreateDefaultRollapp()
	keeper := suite.App.SequencerKeeper

	err := keeper.JailSequencerOnFraud(suite.Ctx, "unknown_sequencer")
	suite.ErrorIs(err, types.ErrUnknownSequencer)
}

func (suite *SequencerTestSuite) TestSlashingUnbondedSequencer() {
	keeper := suite.App.SequencerKeeper
	rollappId, _ := suite.CreateDefaultRollappAndProposer()

	seqAddr := suite.KeeperTestHelper.CreateDefaultSequencer(suite.Ctx, rollappId) // bonded non proposer

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	res, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	unbondTime := res.GetUnbondingCompletionTime()
	keeper.UnbondAllMatureSequencers(suite.Ctx, unbondTime.Add(1*time.Second))

	seq, found := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)

	suite.Equal(seq.Address, seqAddr)
	suite.Equal(seq.Status, types.Unbonded)
	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.ErrorIs(err, types.ErrInvalidSequencerStatus)
}

func (suite *SequencerTestSuite) TestSlashingUnbondingSequencer() {
	keeper := suite.App.SequencerKeeper

	rollappId, pk := suite.CreateDefaultRollapp()
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk)

	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId, ed25519.GenPrivKey().PubKey())

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Equal(seq.Status, types.Unbonding)
	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.NoError(err)

	suite.assertSlashed(seqAddr)
}

func (suite *SequencerTestSuite) TestSlashingProposerSequencer() {
	keeper := suite.App.SequencerKeeper

	rollappId, pk1 := suite.CreateDefaultRollapp()
	pk2 := ed25519.GenPrivKey().PubKey()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk1)
	seqAddr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk2)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Equal(seq.Status, types.Bonded)

	seq2, ok := keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Equal(seq2.Status, types.Bonded)

	proposer, _ := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Equal(proposer.Address, seqAddr)

	err := keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.NoError(err)

	suite.assertSlashed(seqAddr)

	seq2, ok = keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Equal(seq2.Status, types.Bonded)

	proposer, found := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().False(found)
}

func (suite *SequencerTestSuite) TestSlashAndJailBondReducingSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId, pk := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	reduceBondMsg := types.MsgDecreaseBond{Creator: seqAddr, DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &reduceBondMsg)
	suite.Require().NoError(err)
	bondReductions := keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 1)

	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.NoError(err)

	bondReductions = keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 0)
	suite.assertSlashed(seqAddr)
}
