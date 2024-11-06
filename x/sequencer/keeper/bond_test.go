package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type DummyBlocker struct {
	called bool
}

func (b *DummyBlocker) CanUnbond(ctx sdk.Context, sequencer types.Sequencer) error {
	b.called = true
	return nil
}

// simple check that blocker is wired in
func (s *SequencerTestSuite) TestBondBlockers() {
	ra := s.createRollapp()
	seq := s.createSequencerWithBond(s.Ctx, ra.RollappId, alice, bond)
	s.k().SetProposer(s.Ctx, ra.RollappId, pkAddr(bob)) // ensure alice is not proposer
	db := DummyBlocker{}
	s.k().SetUnbondBlockers(&db)
	_ = s.k().TryUnbond(s.Ctx, &seq, seq.TokensCoin())
	s.Require().True(db.called)
}
