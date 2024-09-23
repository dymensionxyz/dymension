package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (suite *KeeperTestSuite) TestReplaceDistrRecords() {
	distrInfo, err := types.NewDistrInfo(defaultDistrInfo)
	suite.Require().NoError(err)

	initialWeight := distrInfo.TotalWeight

	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		expectErr          bool
		expectTotalWeight  sdk.Int
		streamId           uint64
	}{
		{
			name: "happy flow - same gauges with different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(10),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(20),
				},
			},
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(30),
		},
		{
			name:               "happy flow - same gauges with same weights",
			testingDistrRecord: defaultDistrInfo,
			expectErr:          false,
			expectTotalWeight:  initialWeight,
		},
		{
			name: "happy flow - changing gauges",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(10),
				},
				{
					GaugeId: 3,
					Weight:  sdk.NewInt(20),
				},
			},
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(30),
		},
		{
			name:               "Not existent stream.",
			testingDistrRecord: defaultDistrInfo,
			expectErr:          true,
			streamId:           12,
		},
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 12,
				Weight:  sdk.NewInt(100),
			}},
			expectErr: true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
				},
			},
			expectErr: true,
		},
		{
			name: "Adding unsort gauges at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(200),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(250),
				},
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()

			err := suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)

			id, _ := suite.CreateDefaultStream(sdk.NewCoins(sdk.NewInt64Coin("udym", 100000)))
			if test.streamId != 0 {
				id = test.streamId
			}

			err = suite.App.StreamerKeeper.ReplaceDistrRecords(suite.Ctx, id, test.testingDistrRecord)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, id)
				suite.Require().NoError(err)

				suite.Require().Equal(len(test.testingDistrRecord), len(stream.DistributeTo.Records))
				suite.Require().Equal(test.expectTotalWeight, stream.DistributeTo.TotalWeight)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUpdateDistrRecords() {
	distrInfo, err := types.NewDistrInfo(defaultDistrInfo)
	suite.Require().NoError(err)

	initialWeight := distrInfo.TotalWeight

	tests := []struct {
		name               string
		testingDistrRecord []types.DistrRecord
		expectErr          bool
		expectTotalWeight  sdk.Int
		streamId           uint64
	}{
		{
			name: "happy flow - same gauges with different weights",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(10),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(20),
				},
			},
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(30),
		},
		{
			name: "happy flow - remove gauge and add new one",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(0),
				},
				{
					GaugeId: 3,
					Weight:  sdk.NewInt(30),
				},
			},
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(80),
		},
		{
			name:               "happy flow - same gauges with same weights",
			testingDistrRecord: defaultDistrInfo,
			expectErr:          false,
			expectTotalWeight:  initialWeight,
		},
		{
			name: "happy flow - changing gauges",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(0),
				},
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(0),
				},
				{
					GaugeId: 3,
					Weight:  sdk.NewInt(20),
				},
			},
			expectErr:         false,
			expectTotalWeight: sdk.NewInt(20),
		},
		{
			name:               "Not existent stream.",
			testingDistrRecord: defaultDistrInfo,
			expectErr:          true,
			streamId:           12,
		},
		{
			name: "Not existent gauge.",
			testingDistrRecord: []types.DistrRecord{{
				GaugeId: 12,
				Weight:  sdk.NewInt(100),
			}},
			expectErr: true,
		},
		{
			name: "Adding two of the same gauge id at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(100),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(200),
				},
			},
			expectErr: true,
		},
		{
			name: "Adding unsort gauges at once should error",
			testingDistrRecord: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  sdk.NewInt(200),
				},
				{
					GaugeId: 1,
					Weight:  sdk.NewInt(250),
				},
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()

			err := suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)
			err = suite.CreateGauge()
			suite.Require().NoError(err)

			id, _ := suite.CreateDefaultStream(sdk.NewCoins(sdk.NewInt64Coin("udym", 100000)))
			if test.streamId != 0 {
				id = test.streamId
			}

			err = suite.App.StreamerKeeper.UpdateDistrRecords(suite.Ctx, id, test.testingDistrRecord)
			if test.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, id)
				suite.Require().NoError(err)
				suite.Require().Equal(test.expectTotalWeight, stream.DistributeTo.TotalWeight)
			}
		})
	}
}
