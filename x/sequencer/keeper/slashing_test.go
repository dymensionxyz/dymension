package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestSlashBasic() {
	s.Run("slash at zero does not error", func() {
		// There shouldn't be an error if the sequencer has no tokens
		k := s.App.SequencerKeeper
		rollappId, pk := s.CreateDefaultRollapp()
		seqAddr := s.CreateDefaultSequencer(s.Ctx, rollappId, pk)
		seq, found := k.GetSequencer(s.Ctx, seqAddr)
		s.Require().True(found)
		err := k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
		err = k.Slash(s.Ctx, &seq, seq.Tokens)
		s.Require().NoError(err)
	})
}

func (suite *SequencerTestSuite) TestSlashAndJailBondReducingSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

<<<<<<< HEAD
	rollappId := suite.CreateDefaultRollapp()
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId) // proposer
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
=======
	rollappId, pk := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)
>>>>>>> main

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

<<<<<<< HEAD
	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
=======
	reduceBondMsg := types.MsgDecreaseBond{Creator: seqAddr, DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &reduceBondMsg)
>>>>>>> main
	suite.Require().NoError(err)
	bondReductions := keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 1)

<<<<<<< HEAD
	seq, _ := keeper.GetSequencer(suite.Ctx, seqAddr)
	keeper.UnbondAllMatureSequencers(suite.Ctx, seq.UnbondTime.Add(1*time.Second))
	seq, _ = keeper.GetSequencer(suite.Ctx, seqAddr)

	suite.Equal(seq.SequencerAddress, seqAddr)
	suite.Equal(seq.Status, types.Unbonded)
	err = keeper.Slashing(suite.Ctx, seqAddr)
	suite.ErrorIs(err, types.ErrInvalidSequencerStatus)
}

func (suite *SequencerTestSuite) TestSlashingUnbondingSequencer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId) // proposer
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	unbondMsg := types.MsgUnbond{Creator: seqAddr}
	_, err := suite.msgServer.Unbond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)

	seq, ok := keeper.GetSequencer(suite.Ctx, seqAddr)
	suite.Require().True(ok)
	suite.Equal(types.Unbonding, seq.Status)
	err = keeper.Slashing(suite.Ctx, seqAddr)
=======
	err = keeper.SlashAndJailFraud(suite.Ctx, seqAddr)
>>>>>>> main
	suite.NoError(err)

	bondReductions = keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 0)
	suite.assertSlashed(seqAddr)
}
<<<<<<< HEAD

func (suite *SequencerTestSuite) TestSlashingProposer() {
	suite.SetupTest()
	keeper := suite.App.SequencerKeeper

	rollappId := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	seqAddr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	seq, ok := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(ok)
	suite.Equal(seq.SequencerAddress, seqAddr)

	seq2, ok := keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Equal(seq2.Status, types.Bonded)

	err := keeper.Slashing(suite.Ctx, seqAddr)
	suite.NoError(err)

	suite.assertSlashed(seqAddr)

	seq2, ok = keeper.GetSequencer(suite.Ctx, seqAddr2)
	suite.Require().True(ok)
	suite.Equal(seq2.Status, types.Bonded)

	_, ok = keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)
}
=======
>>>>>>> main
