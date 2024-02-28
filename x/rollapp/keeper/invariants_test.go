package keeper_test

import (
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
)

func (suite *RollappTestSuite) TestInvariants() {
	keeper := suite.app.RollappKeeper
	ctx := suite.ctx

	numOfRollapps := 10
	//create rollapps
	// rollappsuite.CreateDefaultRollapp()
	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()
		keeper.SetRollapp(ctx, rollapp)

		//create sequncers for each rollapp

		//send state updates

		//progress finalization queue
	}

	// check invariant
	suite.Require().NotPanics(func() {
		eibckeeper.DemandOrderCountInvariant(suite.App.EIBCKeeper)(ctx)
		eibckeeper.UnderlyingPacketExistInvariant(suite.App.EIBCKeeper)(ctx)
	})
}
