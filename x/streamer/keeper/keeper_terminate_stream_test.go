package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

var createActiveStreamFunc = func(suite *KeeperTestSuite) uint64 {
	id, stream := suite.CreateStream(defaultDistrInfo, sdk.Coins{sdk.NewInt64Coin("stake", 100)}, time.Time{}, "day", 30)
	err := suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)
	return id
}

var createUpcomingStreamFunc = func(suite *KeeperTestSuite) uint64 {
	id, _ := suite.CreateStream(defaultDistrInfo, sdk.Coins{sdk.NewInt64Coin("stake", 100)}, time.Now().Add(10*time.Minute), "day", 30)
	_, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, id)
	suite.Require().NoError(err)
	return id
}

var createFinishedStreamFunc = func(suite *KeeperTestSuite) uint64 {
	id, stream := suite.CreateStream(defaultDistrInfo, sdk.Coins{sdk.NewInt64Coin("stake", 100)}, time.Time{}, "day", 30)
	err := suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.MoveActiveStreamToFinishedStream(suite.Ctx, *stream)
	suite.Require().NoError(err)
	return id
}

func (suite *KeeperTestSuite) TestTerminateStream() {
	tests := []struct {
		name             string
		expectErr        bool
		createStreamFunc func(suite *KeeperTestSuite) uint64
	}{
		{
			name:             "stop active stream",
			expectErr:        false,
			createStreamFunc: createActiveStreamFunc,
		},
		{
			name:             "stop upcoming stream",
			expectErr:        false,
			createStreamFunc: createUpcomingStreamFunc,
		},
		{
			name:             "stop already finished stream",
			expectErr:        true,
			createStreamFunc: createFinishedStreamFunc,
		},
		{
			name:             "stop non-existent stream",
			expectErr:        true,
			createStreamFunc: func(_ *KeeperTestSuite) uint64 { return 1000 },
		},
	}

	for _, tc := range tests {

		id := tc.createStreamFunc(suite)

		if tc.expectErr {
			err := suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id)
			suite.Require().Error(err, tc.name)
		} else {
			_, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, id)
			suite.Require().NoError(err, tc.name)
			err = suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id)
			suite.Require().NoError(err, tc.name)
		}
	}
}

// test GetModuleToDistributeCoins
func (suite *KeeperTestSuite) TestTerminateStream_ModuleToDistributeCoins() {
	coinUpcoming := sdk.NewInt64Coin("stake", 10000)
	coinActive := coinUpcoming

	// create upcoming stream
	id1, _ := suite.CreateStream(defaultDistrInfo, sdk.Coins{coinUpcoming}, time.Now().Add(10*time.Minute), "day", 30)

	// create active stream
	id2, stream2 := suite.CreateStream(defaultDistrInfo, sdk.Coins{coinActive}, time.Time{}, "day", 30)
	err := suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream2)
	suite.Require().NoError(err)

	toDist := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coinUpcoming.Add(coinActive).Amount, toDist.AmountOf("stake"))

	// stop streams
	err = suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id1)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id2)
	suite.Require().NoError(err)

	toDist = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(sdk.ZeroInt(), toDist.AmountOf("stake"))
}

// test GetModuleDistributedCoins
func (suite *KeeperTestSuite) TestTerminateStream_ModuleDistributedCoins() {
	coinUpcoming := sdk.NewInt64Coin("stake", 10000)
	coinActive := coinUpcoming

	err := suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)

	// create upcoming stream
	id1, _ := suite.CreateStream(defaultDistrInfo, sdk.Coins{coinUpcoming}, time.Now().Add(10*time.Minute), "day", 30)

	// create active stream
	id2, stream2 := suite.CreateStream(defaultDistrInfo, sdk.Coins{coinActive}, time.Time{}, "day", 10)
	err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream2)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.UpdateStreamAtEpochStart(suite.Ctx, *stream2)
	suite.Require().NoError(err)

	toDist := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coinUpcoming.Add(coinActive).Amount, toDist.AmountOf("stake"))

	// stop streams
	err = suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id1)
	suite.Require().NoError(err)
	err = suite.App.StreamerKeeper.TerminateStream(suite.Ctx, id2)
	suite.Require().NoError(err)

	/* ---------------------- check ModuleToDistributeCoins --------------------- */
	toDist = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(sdk.ZeroInt(), toDist.AmountOf("stake"))

	distributed, err := suite.App.StreamerKeeper.AfterEpochEnd(suite.Ctx, "day")
	suite.Require().NoError(err)
	suite.Require().Empty(distributed)

	/* ---------------------- check ModuleDistributedCoins ---------------------- */
	expectedDist := suite.App.StreamerKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(distributed.AmountOf("stake"), expectedDist.AmountOf("stake"))
}
