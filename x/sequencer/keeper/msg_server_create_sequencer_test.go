package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

// On success, we should get back an object with all the right info
func (s *SequencerTestSuite) TestCreateSequencerBasic() {
	ra := s.createRollapp()
	s.fundSequencer(alice, bond)
	msg := createSequencerMsg(ra.RollappId, alice)
	msg.Bond = bond
	_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
	s.Require().NoError(err)
	seq, err := s.k().GetRealSequencer(s.Ctx, pkAddr(alice))
	s.Require().NoError(err)
	s.Require().Equal(seq.Address, pkAddr(alice))
	s.Require().True(bond.Equal(seq.TokensCoin()))
	// TODO: check balance is reduced
}

// There are several reasons to reject creation
func (s *SequencerTestSuite) TestCreateSequencerRestrictions() {
	ra := s.createRollapp()

	s.Run("no rollapp", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(urand.RollappID(), alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("already exist", func() {
		pk := randPK()
		s.fundSequencer(pk, bond)
		msg := createSequencerMsg(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})
	s.Run("TODO: awaitingLastProposerBlock", func() {
	})
	s.Run("insufficient bond", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Amount = msg.Bond.Amount.Sub(sdk.OneInt())
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrOutOfRange)
	})
	s.Run("wrong denom", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Denom = "foo"
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("TODO: vm", func() {
	})
	s.Run("TODO: launched", func() {
	})
}
