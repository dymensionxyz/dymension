package keeper_test

func (suite *RollappTestSuite) TestKeeper_SetRegisteredDenom() {
	denoms := []string{
		"stake",
		"adym",
		"ibc/A88EE35932B15B981676EFA6700342EDEF63C41C9EE1265EA5BEDAE0A6518CEA",
	}

	numRollapps := 3
	rollappIDs := make([]string, 0, numRollapps)

	for i := 0; i < numRollapps; i++ {
		rollappID := suite.CreateDefaultRollapp()
		rollappIDs = append(rollappIDs, rollappID)

		for _, d := range denoms {
			err := suite.App.RollappKeeper.SetRegisteredDenom(suite.Ctx, rollappID, d)
			suite.Require().NoError(err)
			suite.Require().True(suite.App.RollappKeeper.HasRegisteredDenom(suite.Ctx, rollappID, d))
		}
	}

	for _, rollappId := range rollappIDs {
		gotDenoms, err := suite.App.RollappKeeper.GetAllRegisteredDenoms(suite.Ctx, rollappId)
		suite.Require().NoError(err)
		suite.Require().ElementsMatch(denoms, gotDenoms)
	}
}
