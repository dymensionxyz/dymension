package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestFraudSubmittedHook() {
	s.Ctx = s.Ctx.WithBlockHeight(10)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	keeper := s.App.SequencerKeeper

	rollappId, pk := s.CreateDefaultRollapp()

	numOfSequencers := 5

	// create 5 sequencers for rollapp1
	seqAddrs := make([]string, numOfSequencers)
	seqAddrs[0] = s.CreateSequencerWithBond(s.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)

	for i := 1; i < numOfSequencers; i++ {
		pki := ed25519.GenPrivKey().PubKey()
		seqAddrs[i] = s.CreateSequencer(s.Ctx, rollappId, pki)
	}

	proposer := seqAddrs[0]
	p, found := keeper.GetProposerLegacy(s.Ctx, rollappId)
	s.Require().True(found)
	s.Require().Equal(proposer, p.Address)

	// queue the third sequencer to reduce bond
	decreaseBondMsg := types.MsgDecreaseBond{Creator: seqAddrs[0], DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := s.msgServer.DecreaseBond(s.Ctx, &decreaseBondMsg)
	s.Require().NoError(err)
	bds := keeper.GetMatureDecreasingBondIDs(s.Ctx, resp.GetCompletionTime())
	s.Require().Len(bds, 1)

	err = keeper.RollappHooks().FraudSubmitted(s.Ctx, rollappId, 0, proposer)
	s.Require().NoError(err)

	// check if proposer is slashed
	sequencer, found := keeper.GetSequencer(s.Ctx, proposer)
	s.Require().True(found)
	s.Require().True(sequencer.Jailed)
	s.Require().Equal(sequencer.Status, types.Unbonded)

	// check if other sequencers are unbonded
	for i := 1; i < numOfSequencers; i++ {
		sequencer, found := keeper.GetSequencer(s.Ctx, seqAddrs[i])
		s.Require().True(found)
		s.Require().Equal(sequencer.Status, types.Unbonded)
	}

	// check no proposer is set for the rollapp after fraud
	_, ok := keeper.GetProposerLegacy(s.Ctx, rollappId)
	s.Require().False(ok)
	// check if bond reduction queue is pruned
	bds = keeper.GetMatureDecreasingBondIDs(s.Ctx, resp.GetCompletionTime())
	s.Require().Len(bds, 0)
}
