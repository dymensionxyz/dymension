package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestInvariants() {
	s.SetupTest()

	numOfRollapps := 5
	numOfSequencers := 5

	var rollappToTest string

	// create rollapps and sequencers
	for i := 0; i < numOfRollapps; i++ {
		rollapp, pk := s.CreateDefaultRollapp()

		// create sequencers
		seqAddr := make([]string, numOfSequencers)
		seqAddr[0] = s.CreateSequencer(s.Ctx, rollapp, pk)
		for j := 1; j < numOfSequencers; j++ {
			pki := ed25519.GenPrivKey().PubKey()
			seqAddr[j] = s.CreateSequencer(s.Ctx, rollapp, pki)
		}

	}

	rollappid := rollappToTest
	seqUnbonding := s.App.SequencerKeeper.RollappSequencersByStatus(s.Ctx, rollappid, types.Unbonding)
	s.Require().True(len(seqUnbonding) > 0)

	// Test the test: make sure all status have entries
	seqBonded := s.App.SequencerKeeper.RollappSequencersByStatus(s.Ctx, rollappid, types.Bonded)
	seqUnbonding = s.App.SequencerKeeper.RollappSequencersByStatus(s.Ctx, rollappid, types.Unbonding)
	seqUnbonded := s.App.SequencerKeeper.RollappSequencersByStatus(s.Ctx, rollappid, types.Unbonded)

	if len(seqBonded) == 0 || len(seqUnbonding) == 0 || len(seqUnbonded) == 0 {
		s.T().Fatal("Test setup failed")
	}
	// additional rollapp with no sequencers
	s.CreateDefaultRollapp()

	msg, ok := keeper.AllInvariants(s.App.SequencerKeeper)(s.Ctx)
	s.Require().False(ok, msg)
}
