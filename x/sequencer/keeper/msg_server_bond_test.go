package keeper_test

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

func (s *SequencerTestSuite) TestIncreaseBondBasic() {
	ra := s.createRollapp()
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	m := &types.MsgIncreaseBond{
		Creator:   seq.Address,
		AddAmount: bond,
	}
	expect := bond
	for range 2 {
		s.fundSequencer(seq.MustPubKey(), m.AddAmount)
		_, err := s.msgServer.IncreaseBond(s.Ctx, m)
		s.Require().NoError(err)
		expect = expect.Add(bond)
		seq = s.k().GetSequencer(s.Ctx, seq.Address)
		s.Require().True(expect.Equal(seq.TokensCoin()))
		s.Require().True(expect.Equal(s.moduleBalance()))
	}
}

func (s *SequencerTestSuite) TestIncreaseBondRestrictions() {
	ra := s.createRollapp()

	s.Run("wrong denom", func() {
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
		m := &types.MsgIncreaseBond{
			Creator:   seq.Address,
			AddAmount: bond,
		}
		m.AddAmount.Denom = "foo"
		_, err := s.msgServer.IncreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("sequencer not found", func() {
		// do not create sequencer
		m := &types.MsgIncreaseBond{
			Creator:   pkAddr(bob),
			AddAmount: bond,
		}
		_, err := s.msgServer.IncreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("insufficient funds", func() {
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, bond)
		m := &types.MsgIncreaseBond{
			Creator:   seq.Address,
			AddAmount: bond,
		}
		// do not fund
		_, err := s.msgServer.IncreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, sdkerrors.ErrInsufficientFunds)
	})
}
