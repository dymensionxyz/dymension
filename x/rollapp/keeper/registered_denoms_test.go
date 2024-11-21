package keeper_test

func (s *RollappTestSuite) TestKeeper_SetRegisteredDenom() {
	denoms := []string{
		"stake",
		"adym",
		"ibc/A88EE35932B15B981676EFA6700342EDEF63C41C9EE1265EA5BEDAE0A6518CEA",
	}

	numRollapps := 3
	rollappIDs := make([]string, 0, numRollapps)

	for i := 0; i < numRollapps; i++ {
		rollappID := s.CreateDefaultRollapp()
		rollappIDs = append(rollappIDs, rollappID)

		for _, d := range denoms {
			err := s.k().SetRegisteredDenom(s.Ctx, rollappID, d)
			s.Require().NoError(err)
			s.Require().True(s.k().HasRegisteredDenom(s.Ctx, rollappID, d))
		}
	}

	for _, rollappId := range rollappIDs {
		gotDenoms, err := s.k().GetAllRegisteredDenoms(s.Ctx, rollappId)
		s.Require().NoError(err)
		s.Require().ElementsMatch(denoms, gotDenoms)
	}
}
