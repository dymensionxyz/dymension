package keeper_test

import (
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
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})
}
