package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestInvariants() {
	suite.SetupTest()
	initialheight := uint64(10)
	initialTime := time.Now()

	numOfRollapps := 5
	numOfSequencers := 5

	// create rollapps and sequencers
	for i := 0; i < numOfRollapps; i++ {
		rollapp, pk := suite.CreateDefaultRollapp()

		// create sequencers
		seqAddr := make([]string, numOfSequencers)
		seqAddr[0] = suite.CreateDefaultSequencer(suite.Ctx, rollapp, pk)
		for j := 1; j < numOfSequencers; j++ {
			pki := ed25519.GenPrivKey().PubKey()
			seqAddr[j] = suite.CreateDefaultSequencer(suite.Ctx, rollapp, pki)
		}

		// unbonding some sequencers
		for j := uint64(0); j < uint64(numOfSequencers/2); j++ {
			suite.Ctx = suite.Ctx.WithBlockHeight(int64(initialheight + j)).WithBlockTime(initialTime.Add(time.Duration(j) * time.Second))
			_, err := suite.msgServer.Unbond(suite.Ctx, &types.MsgUnbond{Creator: seqAddr[j]})
			suite.Require().NoError(err)
		}
	}
	// unbond some
	unbondTime := initialTime.Add(suite.App.SequencerKeeper.UnbondingTime(suite.Ctx))
	suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, unbondTime)

	// Test the test: make sure all status have entries
	rollappid := suite.App.RollappKeeper.GetAllRollapps(suite.Ctx)[0].RollappId
	seqBonded := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Bonded)
	seqUnbonding := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Unbonding)
	seqUnbonded := suite.App.SequencerKeeper.GetSequencersByRollappByStatus(suite.Ctx, rollappid, types.Unbonded)

	if len(seqBonded) == 0 || len(seqUnbonding) == 0 || len(seqUnbonded) == 0 {
		suite.T().Fatal("Test setup failed")
	}
	// additional rollapp with no sequencers
	suite.CreateDefaultRollapp()

	msg, ok := keeper.AllInvariants(suite.App.SequencerKeeper)(suite.Ctx)
	suite.Require().False(ok, msg)
}
