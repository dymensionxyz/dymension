package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	seqAddrs[0] = suite.CreateSequencerWithBond(suite.Ctx, rollappId, bond.AddAmount(sdk.NewInt(20)), pk)

	for i := 1; i < numOfSequencers; i++ {
		pki := ed25519.GenPrivKey().PubKey()
		seqAddrs[i] = suite.CreateSequencer(suite.Ctx, rollappId, pki)
	}

	proposer := seqAddrs[0]
	p, found := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().True(found)
	suite.Require().Equal(proposer, p.Address)

	// queue the third sequencer to reduce bond
	unbondMsg := types.MsgDecreaseBond{Creator: seqAddrs[0], DecreaseAmount: sdk.NewInt64Coin(bond.Denom, 10)}
	resp, err := suite.msgServer.DecreaseBond(suite.Ctx, &unbondMsg)
	suite.Require().NoError(err)
	bds := keeper.GetMatureDecreasingBondIDs(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bds, 1)

	err = keeper.RollappHooks().FraudSubmitted(suite.Ctx, rollappId, 0, proposer)
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
		suite.Require().Equal(sequencer.Status, types.Unbonded)
	}

	// check no proposer is set for the rollapp after fraud
	_, ok := keeper.GetProposer(suite.Ctx, rollappId)
	suite.Require().False(ok)
	// check if bond reduction queue is pruned
	bds = keeper.GetMatureDecreasingBondIDs(suite.Ctx, resp.GetCompletionTime())
	suite.Require().Len(bds, 0)
}
