package keeper_test

import "github.com/stretchr/testify/suite"

func (suite *KeeperTestSuite) TestStreamReferencesManagement() {
	key1 := []byte{0x11}
	key2 := []byte{0x12}

	suite.SetupTest()

	// set two stream references to key 1 and three stream references to key 2
	_ = suite.App.StreamerKeeper.AddStreamRefByKey(suite.Ctx, key1, 1)
	_ = suite.App.StreamerKeeper.AddStreamRefByKey(suite.Ctx, key2, 1)
	_ = suite.App.StreamerKeeper.AddStreamRefByKey(suite.Ctx, key1, 2)
	_ = suite.App.StreamerKeeper.AddStreamRefByKey(suite.Ctx, key2, 2)
	_ = suite.App.StreamerKeeper.AddStreamRefByKey(suite.Ctx, key2, 3)

	// ensure key1 only has 2 entires
	streamRefs1 := suite.App.StreamerKeeper.GetStreamRefs(suite.Ctx, key1)
	suite.Require().Equal(len(streamRefs1), 2)

	// ensure key2 only has 3 entries
	streamRefs2 := suite.App.StreamerKeeper.GetStreamRefs(suite.Ctx, key2)
	suite.Require().Equal(len(streamRefs2), 3)

	// remove stream 1 from key2, resulting in a reduction from 3 to 2 entries
	err := suite.App.StreamerKeeper.DeleteStreamRefByKey(suite.Ctx, key2, 1)
	suite.Require().NoError(err)

	// ensure key2 now only has 2 entires
	streamRefs3 := suite.App.StreamerKeeper.GetStreamRefs(suite.Ctx, key2)
	suite.Require().Equal(len(streamRefs3), 2)
}

var _ = suite.TestingSuite(nil)
