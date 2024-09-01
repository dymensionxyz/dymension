package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestJailUnknownSequencer() {
	suite.CreateDefaultRollapp()
	keeper := suite.App.SequencerKeeper

	err := keeper.JailSequencerOnFraud(suite.Ctx, "unknown_sequencer")
	suite.ErrorIs(err, types.ErrUnknownSequencer)
}

func (suite *SequencerTestSuite) TestJailUnbondedSequencer() {
	keeper := suite.App.SequencerKeeper
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	rollappId, _ := suite.CreateDefaultRollappAndProposer()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId) // bonded non proposer

	// unbond the non-proposer
	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	res, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)
	unbondTime := res.GetUnbondingCompletionTime()
	keeper.UnbondAllMatureSequencers(suite.Ctx, unbondTime.Add(1*time.Second))
	seq, found := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(found)
	suite.Equal(seq.Address, seqAddr)
	suite.Equal(seq.Status, types.Unbonded)

	// jail the unbonded sequencer
	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.ErrorIs(err, types.ErrInvalidSequencerStatus)
}

func (suite *SequencerTestSuite) TestJailUnbondingSequencer() {
	keeper := suite.App.SequencerKeeper
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	rollappId, _ := suite.CreateDefaultRollappAndProposer()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId) // bonded non proposer

	// unbond the non-proposer
	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)
	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Equal(seq.Status, types.Unbonding)

	// jail the unbonding sequencer
	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.NoError(err)
	suite.assertJailed(seqAddr)
}

func (suite *SequencerTestSuite) TestJailProposerSequencer() {
	keeper := suite.App.SequencerKeeper
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	rollappId, proposer := suite.CreateDefaultRollappAndProposer()
	err := keeper.JailSequencerOnFraud(suite.Ctx, proposer)
	suite.NoError(err)
	suite.assertJailed(proposer)

	_, found := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().False(found)
}

func (suite *SequencerTestSuite) TestJailBondReducingSequencer() {
	keeper := suite.App.SequencerKeeper
	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	rollappId, pk := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)

	reduceBondMsg := types.MsgDecreaseBond{Creator: seqAddr, DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &reduceBondMsg)
	suite.Require().NoError(err)
	bondReductions := keeper.GetMatureDecreasingBondIDs(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 1)

	err = keeper.JailSequencerOnFraud(suite.Ctx, seqAddr)
	suite.NoError(err)

	bondReductions = keeper.GetMatureDecreasingBondIDs(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 0)
	suite.assertJailed(seqAddr)
}
