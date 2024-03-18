package keeper_test

import (
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func (suite *RollappTestSuite) TestInvariants() {
	suite.SetupTest()
	ctx := suite.Ctx
	initialheight := int64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight)

	numOfRollapps := 10
	numOfStates := 10

	// create rollapps
	seqPerRollapp := make(map[string]string)
	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()
		seqaddr := suite.CreateDefaultSequencer(ctx, rollapp)

		// skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
	}

	rollapp := suite.CreateRollappWithName("dym_1100-1")
	seqaddr := suite.CreateDefaultSequencer(ctx, rollapp)
	seqPerRollapp[rollapp] = seqaddr

	rollapp = suite.CreateRollappWithName("dym_1100")
	seqaddr = suite.CreateDefaultSequencer(ctx, rollapp)
	seqPerRollapp[rollapp] = seqaddr

	// send state updates
	var lastHeight uint64 = 0
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp := range seqPerRollapp {
			_, err := suite.PostStateUpdate(ctx, rollapp, seqPerRollapp[rollapp], lastHeight+1, numOfBlocks)
			suite.Require().Nil(err)
		}

		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
		lastHeight = lastHeight + numOfBlocks
	}

	// progress finalization queue
	// disputePeriod := int64(suite.App.RollappKeeper.GetParams(ctx).DisputePeriodInBlocks)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight + 2)
	suite.App.RollappKeeper.FinalizeQueue(suite.Ctx)

	// check invariant
	msg, bool := keeper.AllInvariants(suite.App.RollappKeeper)(suite.Ctx)
	suite.Require().False(bool, msg)
}
