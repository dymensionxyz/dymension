package keeper_test

import (
	collections "cosmossdk.io/collections"
)

func (suite *KeeperTestSuite) TestLPs() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx

	err = k.LPs.M.Set(ctx, 1, 1)
	suite.Require().NoError(err)
	err = k.LPs.M.Set(ctx, 3, 1)
	suite.Require().NoError(err)
	err = k.LPs.M.Set(ctx, 5, 1)

	//var rng collections.Ranger[uint64]
	//rng := collections.NewPrefixUntilPairRange()
	rng := new(collections.Range[uint64]).StartInclusive(2)
	iter, err := k.LPs.M.Iterate(ctx, rng)
	suite.Require().NoError(err)
	keys, err := iter.Keys()
	suite.Require().NoError(err)
	for _, x := range keys {
		suite.T().Log(x)
	}
}
