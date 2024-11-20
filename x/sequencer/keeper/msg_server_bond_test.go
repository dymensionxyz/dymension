package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

func (s *SequencerTestSuite) TestIncreaseBondBasic() {
	ra := s.createRollapp()
	expect := bond
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, expect)
	m := &types.MsgIncreaseBond{
		Creator:   seq.Address,
		AddAmount: bond,
	}
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

func (s *SequencerTestSuite) TestDecreaseBondBasic() {
	ra := s.createRollapp()
	expect := ucoin.SimpleMul(bond, 10) // plenty
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, expect)
	m := &types.MsgDecreaseBond{
		Creator:        seq.Address,
		DecreaseAmount: bond,
	}
	s.k().SetProposer(s.Ctx, ra.RollappId, pkAddr(randomTMPubKey())) // make not proposer so it's allowed
	for range 2 {
		_, err := s.msgServer.DecreaseBond(s.Ctx, m)
		s.Require().NoError(err)
		expect = expect.Sub(bond)
		seq = s.k().GetSequencer(s.Ctx, seq.Address)
		s.Require().True(expect.Equal(seq.TokensCoin()))
		s.Require().True(expect.Equal(s.moduleBalance()))
	}
}

func (s *SequencerTestSuite) TestDecreaseBondRestrictions() {
	ra := s.createRollapp()

	s.Run("sequencer not found", func() {
		// do not create sequencer
		m := &types.MsgDecreaseBond{
			Creator:        pkAddr(alice),
			DecreaseAmount: bond,
		}
		_, err := s.msgServer.DecreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("proposer", func() {
		currBond := ucoin.SimpleMul(bond, 3)
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, currBond)
		s.k().SetProposer(s.Ctx, ra.RollappId, seq.Address)
		m := &types.MsgDecreaseBond{
			Creator:        seq.Address,
			DecreaseAmount: bond,
		}
		_, err := s.msgServer.DecreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})
	s.Run("successor", func() {
		currBond := ucoin.SimpleMul(bond, 3)
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, currBond)
		s.k().SetSuccessor(s.Ctx, ra.RollappId, seq.Address)
		m := &types.MsgDecreaseBond{
			Creator:        seq.Address,
			DecreaseAmount: bond,
		}
		_, err := s.msgServer.DecreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})
	s.Run("fall below min", func() {
		currBond := ucoin.MulDec(sdk.MustNewDecFromStr("1.5"), bond)[0]
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, pks[3], currBond)
		m := &types.MsgDecreaseBond{
			Creator:        seq.Address,
			DecreaseAmount: bond, // too much
		}
		_, err := s.msgServer.DecreaseBond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})
}

func (s *SequencerTestSuite) TestUnbondBasic() {
	ra := s.createRollapp()
	expect := bond
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, expect)
	s.k().SetProposer(s.Ctx, ra.RollappId, pkAddr(randomTMPubKey())) // make not proposer so it's allowed
	m := &types.MsgUnbond{
		Creator: seq.Address,
	}
	_, err := s.msgServer.Unbond(s.Ctx, m)
	s.Require().NoError(err)
	seq = s.k().GetSequencer(s.Ctx, seq.Address)
	s.Require().Equal(types.Unbonded, seq.Status)
	s.Require().True(s.moduleBalance().IsZero())
	s.Require().True(seq.TokensCoin().IsZero())
}

func (s *SequencerTestSuite) TestUnbondRestrictions() {
	ra := s.createRollapp()

	s.Run("sequencer not found", func() {
		// do not create sequencer
		m := &types.MsgUnbond{
			Creator: pkAddr(alice),
		}
		_, err := s.msgServer.Unbond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("proposer - start notice", func() {
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, bob, bond)
		s.k().SetProposer(s.Ctx, ra.RollappId, seq.Address)
		m := &types.MsgUnbond{
			Creator: seq.Address,
		}
		res, err := s.msgServer.Unbond(s.Ctx, m)
		s.Require().NoError(err)
		s.Require().False(res.GetNoticePeriodCompletionTime().IsZero())
		seq = s.k().GetSequencer(s.Ctx, seq.Address)
		s.Require().True(seq.NoticeInProgress(s.Ctx.BlockTime()))
		s.Require().True(s.k().IsProposer(s.Ctx, seq))
		s.Require().False(seq.OptedIn)
	})
	s.Run("successor - not allowed", func() {
		seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, charlie, bond)
		s.k().SetSuccessor(s.Ctx, ra.RollappId, seq.Address)
		m := &types.MsgUnbond{
			Creator: seq.Address,
		}
		_, err := s.msgServer.Unbond(s.Ctx, m)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})
}
