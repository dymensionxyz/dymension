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

	rollappId, pk := suite.CreateDefaultRollapp()
	seqAddr := suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)

	suite.Ctx = suite.Ctx.WithBlockHeight(20)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	reduceBondMsg := types.MsgDecreaseBond{Creator: seqAddr, DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &reduceBondMsg)
	suite.Require().NoError(err)
	bondReductions := keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 1)

	err = keeper.SlashAndJailFraud(suite.Ctx, seqAddr)
	suite.NoError(err)

	bondReductions = keeper.GetMatureDecreasingBondSequencers(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bondReductions, 0)
	suite.assertSlashed(seqAddr)
}
