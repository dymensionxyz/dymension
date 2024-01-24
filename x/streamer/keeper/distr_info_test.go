package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) TestAllocateToGauges() {
	tests := []struct {
		name                   string
		testingDistrRecord     []types.DistrRecord
		mintedCoins            sdk.Coin
		expectedGaugesBalances []sdk.Coins
		expectedCommunityPool  sdk.DecCoin
	}{
		// With minting 15000 stake to module, after AllocateAsset we get:
		// expectedCommunityPool = 0 (All reward will be transferred to the gauges)
		// 	expectedGaugesBalances in order:
		//    gaue1_balance = 15000 * 100/(100+200+300) = 2500
		//    gaue2_balance = 15000 * 200/(100+200+300) = 5000 (using the formula in the function gives the exact result 4999,9999999999995000. But TruncateInt return 4999. Is this the issue?)
		//    gaue3_balance = 15000 * 300/(100+200+300) = 7500
		{
			name: "Allocated to the gauges proportionally",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(200),
				},
				{
					GaugeId: 3,
					Weight:  sdk.NewInt(300),
				},
			},
			mintedCoins: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(15000)),
			expectedGaugesBalances: []sdk.Coins{
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4999))),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(7500))),
			},
			expectedCommunityPool: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
		},
	}

	for name, test := range tests {
		suite.Run(test.name, func() {

			var streams []types.Stream
			suite.SetupTest()

			err := suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)

			keeper := suite.App.StreamerKeeper

			// create a stream
			suite.CreateStream(test.testingDistrRecord, sdk.NewCoins(test.mintedCoins), time.Now(), "day", 1)

			// move all created streams from upcoming to active
			suite.Ctx = suite.Ctx.WithBlockTime(time.Now())
			streams = suite.App.StreamerKeeper.GetStreams(suite.Ctx)
			for _, stream := range streams {
				err := suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, stream)
				suite.Require().NoError(err)
			}

			_, err = keeper.Distribute(suite.Ctx, streams)
			suite.Require().NoError(err, name)

			for i := 0; i < len(test.testingDistrRecord); i++ {
				if test.testingDistrRecord[i].GaugeId == 0 {
					continue
				}
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, test.testingDistrRecord[i].GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(test.expectedGaugesBalances[i], gauge.Coins)
			}
		})
	}
}

func TestNewDistrInfo(t *testing.T) {
	// Test case: valid records
	records := []types.DistrRecord{
		{Weight: sdk.NewInt(1)},
		{Weight: sdk.NewInt(2)},
	}
	distrInfo, err := types.NewDistrInfo(records)
	require.NoError(t, err)
	require.Equal(t, distrInfo.TotalWeight, sdk.NewInt(3))

	// Test case: invalid record
	records = []types.DistrRecord{
		{Weight: sdk.NewInt(-1)},
	}
	distrInfo, err = types.NewDistrInfo(records)
	require.Error(t, err)

	// Test case: total weight not positive
	records = []types.DistrRecord{}
	distrInfo, err = types.NewDistrInfo(records)
	require.Error(t, err)
	require.Equal(t, err, types.ErrDistrInfoNotPositiveWeight)
}
