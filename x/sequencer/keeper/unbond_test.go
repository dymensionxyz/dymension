package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestUnbondingMultiple() {
	s.Ctx = s.Ctx.WithBlockHeight(10)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	keeper := s.App.SequencerKeeper

	rollappId, pk1 := s.CreateDefaultRollapp()
	rollappId2, pk2 := s.CreateDefaultRollapp()

	numOfSequencers := 4
	numOfSequencers2 := 3
	unbodingSeq := 2

	seqAddr1 := make([]string, numOfSequencers)
	seqAddr2 := make([]string, numOfSequencers2)

	// create 5 sequencers for rollapp1
	seqAddr1[0] = s.CreateSequencer(s.Ctx, rollappId, pk1)
	for i := 1; i < numOfSequencers; i++ {
		seqAddr1[i] = s.CreateSequencer(s.Ctx, rollappId, ed25519.GenPrivKey().PubKey())
	}

	// create 3 sequencers for rollapp2
	seqAddr2[0] = s.CreateSequencer(s.Ctx, rollappId2, pk2)
	for i := 1; i < numOfSequencers2; i++ {
		seqAddr2[i] = s.CreateSequencer(s.Ctx, rollappId2, ed25519.GenPrivKey().PubKey())
	}

	// start unbonding for 2 sequencers in each rollapp
	s.Ctx = s.Ctx.WithBlockHeight(20)
	now := time.Now()
	unbondTime := now.Add(keeper.GetParams(s.Ctx).UnbondingTime)
	s.Ctx = s.Ctx.WithBlockTime(now)
	for i := 1; i < unbodingSeq+1; i++ {
		unbondMsg := types.MsgUnbond{Creator: seqAddr1[i]}
		_, err := s.msgServer.Unbond(s.Ctx, &unbondMsg)
		s.Require().NoError(err)

		unbondMsg = types.MsgUnbond{Creator: seqAddr2[i]}
		_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
		s.Require().NoError(err)
	}

	// before unbonding time reached
	sequencers := keeper.GetMatureUnbondingSequencers(s.Ctx, now)
	s.Require().Len(sequencers, 0)

	sequencers = keeper.GetMatureUnbondingSequencers(s.Ctx, unbondTime.Add(-1*time.Second))
	s.Require().Len(sequencers, 0)

	// past unbonding time
	sequencers = keeper.GetMatureUnbondingSequencers(s.Ctx, unbondTime.Add(1*time.Second))
	s.Require().Len(sequencers, 4)
}

func (s *SequencerTestSuite) TestTokensRefundOnUnbond() {
	denom := bond.Denom
	blockheight := 20
	var err error

	rollappId, pk := s.CreateDefaultRollapp()
	_ = s.CreateSequencer(s.Ctx, rollappId, pk)

	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := s.CreateSequencer(s.Ctx, rollappId, pk1)
	sequencer1, _ := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr1)
	s.Require().True(sequencer1.Status == types.Bonded)
	s.Require().False(sequencer1.Tokens.IsZero())

	pk2 := ed25519.GenPrivKey().PubKey()
	addr2 := s.CreateSequencer(s.Ctx, rollappId, pk2)
	sequencer2, _ := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr2)
	s.Require().True(sequencer2.Status == types.Bonded)
	s.Require().False(sequencer2.Tokens.IsZero())

	s.Ctx = s.Ctx.WithBlockHeight(int64(blockheight))
	s.Ctx = s.Ctx.WithBlockTime(time.Now())

	// start the 1st unbond
	unbondMsg := types.MsgUnbond{Creator: addr1}
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	sequencer1, _ = s.App.SequencerKeeper.GetSequencer(s.Ctx, addr1)
	s.Require().True(sequencer1.Status == types.Unbonding)
	s.Require().Equal(sequencer1.UnbondRequestHeight, int64(blockheight))
	s.Require().False(sequencer1.Tokens.IsZero())

	// start the 2nd unbond later
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(5 * time.Minute))
	unbondMsg = types.MsgUnbond{Creator: addr2}
	_, err = s.msgServer.Unbond(s.Ctx, &unbondMsg)
	s.Require().NoError(err)
	sequencer2, _ = s.App.SequencerKeeper.GetSequencer(s.Ctx, addr2)
	s.Require().True(sequencer2.Status == types.Unbonding)
	s.Require().False(sequencer2.Tokens.IsZero())

	/* -------------------------- check the unbond phase ------------------------- */
	balanceBefore := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(addr1), denom)
	s.App.SequencerKeeper.UnbondAllMatureSequencers(s.Ctx, sequencer1.UnbondTime.Add(1*time.Second))
	balanceAfter := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(addr1), denom)

	// Check stake refunded
	sequencer1, _ = s.App.SequencerKeeper.GetSequencer(s.Ctx, addr1)
	s.Equal(types.Unbonded, sequencer1.Status)
	s.True(sequencer1.Tokens.IsZero())
	s.True(balanceBefore.Add(bond).IsEqual(balanceAfter), "expected %s, got %s", balanceBefore.Add(bond), balanceAfter)

	// check the 2nd unbond still not happened
	sequencer2, _ = s.App.SequencerKeeper.GetSequencer(s.Ctx, addr2)
	s.Equal(types.Unbonding, sequencer2.Status)
	s.False(sequencer2.Tokens.IsZero())
}
