package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestFraudSubmittedHook() {
	suite.Ctx = suite.Ctx.WithBlockHeight(10)
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())

	keeper := suite.App.SequencerKeeper

	rollappId, pk := suite.CreateDefaultRollapp()

	numOfSequencers := 5

	// create 5 sequencers for rollapp1
	seqAddrs := make([]string, numOfSequencers)
	seqAddrs[0] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pk)

	for i := 1; i < numOfSequencers; i++ {
		pki := ed25519.GenPrivKey().PubKey()
		seqAddrs[i] = suite.CreateDefaultSequencer(suite.Ctx, rollappId, pki)
	}
	proposer := seqAddrs[0]

	err := keeper.RollappHooks().FraudSubmitted(suite.Ctx, rollappId, 0, proposer)
	suite.Require().NoError(err)

	// check if proposer is slashed
	sequencer, found := keeper.GetSequencer(suite.Ctx, proposer)
	suite.Require().True(found)
	suite.Require().True(sequencer.Jailed)
	suite.Require().Equal(sequencer.Status, types.Unbonded)

	// check if other sequencers are unbonded
	for i := 1; i < numOfSequencers; i++ {
		sequencer, found := keeper.GetSequencer(suite.Ctx, seqAddrs[i])
		suite.Require().True(found)
		suite.Require().False(sequencer.Proposer)
		suite.Require().Equal(sequencer.Status, types.Unbonded)
	}
}
