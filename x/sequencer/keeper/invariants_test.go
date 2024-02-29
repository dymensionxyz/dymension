package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *SequencerTestSuite) TestInvariants() {
	suite.SetupTest()
	ctx := suite.Ctx
	initialheight := int64(10)
	initialTime := time.Now()
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight).WithBlockTime(initialTime)

	numOfRollapps := 5
	numOfSequencers := 5

	//create rollapps and sequencers
	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()

		//create sequencers
		seqAddr := make([]string, numOfSequencers)
		for j := 0; j < numOfSequencers; j++ {
			seqAddr[j] = suite.CreateDefaultSequencer(ctx, rollapp)
		}
		//unbonding some sequencers
		for j := 0; j < numOfSequencers/2; j++ {
			suite.msgServer.Unbond(ctx, &types.MsgUnbond{seqAddr[j]})
		}

		//unbond some
		unbondTime := initialTime.Add(suite.App.SequencerKeeper.UnbondingTime(suite.Ctx))
		suite.App.SequencerKeeper.UnbondAllMatureSequencers(suite.Ctx, unbondTime.Add(1*time.Second))
	}

	//TODO: make sure all status have entries

	msg, bool := keeper.AllInvariants(suite.App.SequencerKeeper)(suite.Ctx)
	suite.Require().False(bool, msg)
}
