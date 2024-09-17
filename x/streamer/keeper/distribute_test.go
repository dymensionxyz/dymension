package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestDistribute() {
	tests := []struct {
		name    string
		streams []struct {
			coins       sdk.Coins
			numOfEpochs uint64
			distrInfo   []types.DistrRecord
		}
	}{
		{
			name: "single stream single coin",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   []types.DistrRecord
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100)}, 30, defaultDistrInfo}},
		},
		{
			name: "single stream multiple coins",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   []types.DistrRecord
			}{{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, defaultDistrInfo}},
		},
		{
			name: "multiple streams multiple coins multiple epochs",
			streams: []struct {
				coins       sdk.Coins
				numOfEpochs uint64
				distrInfo   []types.DistrRecord
			}{
				{sdk.Coins{sdk.NewInt64Coin("stake", 100), sdk.NewInt64Coin("udym", 300)}, 30, defaultDistrInfo},
				{sdk.Coins{sdk.NewInt64Coin("stake", 1000)}, 365, defaultDistrInfo},
				{sdk.Coins{sdk.NewInt64Coin("udym", 1000)}, 730, defaultDistrInfo},
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// Setup streams and defined in the above tests, then distribute to them

			var streams []types.Stream
			gaugesExpectedRewards := make(map[uint64]sdk.Coins)
			for _, stream := range tc.streams {
				// Create a stream, move it from upcoming to active and update its parameters
				_, newStream := suite.CreateStream(stream.distrInfo, stream.coins, time.Now().Add(-time.Minute), "day", stream.numOfEpochs)

				streams = append(streams, *newStream)

				// Calculate expected rewards
				for _, coin := range stream.coins {
					epochAmt := coin.Amount.Quo(sdk.NewInt(int64(stream.numOfEpochs)))
					if !epochAmt.IsPositive() {
						continue
					}
					for _, record := range newStream.DistributeTo.Records {
						expectedAmtFromStream := epochAmt.Mul(record.Weight).Quo(newStream.DistributeTo.TotalWeight)
						expectedCoins := sdk.Coin{Denom: coin.Denom, Amount: expectedAmtFromStream}
						gaugesExpectedRewards[record.GaugeId] = gaugesExpectedRewards[record.GaugeId].Add(expectedCoins)
					}
				}
			}

			// Trigger the distribution
			suite.DistributeAllRewards(streams)

			// Check expected rewards against actual rewards received
			gauges := suite.App.IncentivesKeeper.GetGauges(suite.Ctx)
			suite.Require().Equal(len(gaugesExpectedRewards), len(gauges), tc.name)
			for _, gauge := range gauges {
				suite.Require().ElementsMatch(gaugesExpectedRewards[gauge.Id], gauge.Coins, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSponsoredDistribute() {
	addrs := apptesting.CreateRandomAccounts(3)

	type stream struct {
		coins       sdk.Coins
		numOfEpochs uint64
		distrInfo   []types.DistrRecord
	}

	tests := []struct {
		name   string
		stream stream
		// true if x/sponsorship has the initial distr. otherwise, the distr if empty
		hasInitialDistr bool
		// the vote that forms the initial distribution
		initialVote sponsorshiptypes.MsgVote
		// true if the x/sponsorship distr has changed between the stream creation and distribution.
		// simulation of the distribution update at the middle of the epoch.
		hasIntermediateDistr bool
		// the vote that forms the intermediate distribution
		intermediateVote sponsorshiptypes.MsgVote
		// is the epoch filled as a side effect
		fillEpochs bool
	}{
		{
			name: "single-coin stream, no initial nor intermediate distributions",
			stream: stream{
				coins:       sdk.Coins{sdk.NewInt64Coin("stake", 100)},
				numOfEpochs: 30,
				distrInfo:   defaultDistrInfo,
			},
			hasInitialDistr:      false,
			initialVote:          sponsorshiptypes.MsgVote{},
			hasIntermediateDistr: false,
			intermediateVote:     sponsorshiptypes.MsgVote{},
			fillEpochs:           false,
		},
		{
			name: "single-coin stream, initial distribution",
			stream: stream{
				coins:       sdk.Coins{sdk.NewInt64Coin("stake", 100)},
				numOfEpochs: 30,
				distrInfo:   defaultDistrInfo,
			},
			hasInitialDistr: true,
			initialVote: sponsorshiptypes.MsgVote{
				Voter: addrs[0].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(50)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(30)},
				},
			},
			hasIntermediateDistr: false,
			intermediateVote:     sponsorshiptypes.MsgVote{},
			fillEpochs:           true,
		},
		{
			name: "single-coin stream, intermediate distribution",
			stream: stream{
				coins:       sdk.Coins{sdk.NewInt64Coin("stake", 100)},
				numOfEpochs: 30,
				distrInfo:   defaultDistrInfo,
			},
			hasInitialDistr:      false,
			initialVote:          sponsorshiptypes.MsgVote{},
			hasIntermediateDistr: true,
			intermediateVote: sponsorshiptypes.MsgVote{
				Voter: addrs[1].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(10)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(90)},
				},
			},
			fillEpochs: true,
		},
		{
			name: "single-coin stream, initial and intermediate distributions",
			stream: stream{
				coins:       sdk.Coins{sdk.NewInt64Coin("stake", 100)},
				numOfEpochs: 30,
				distrInfo:   defaultDistrInfo,
			},
			hasInitialDistr: true,
			initialVote: sponsorshiptypes.MsgVote{
				Voter: addrs[0].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(70)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(30)},
				},
			},
			hasIntermediateDistr: true,
			intermediateVote: sponsorshiptypes.MsgVote{
				Voter: addrs[1].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(10)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(90)},
				},
			},
			fillEpochs: true,
		},
		{
			name: "stream distr info doesn't play any role",
			stream: stream{
				coins:       sdk.Coins{sdk.NewInt64Coin("stake", 100)},
				numOfEpochs: 30,
				// Random unrealistic values
				distrInfo: []types.DistrRecord{
					{
						GaugeId: 121424,
						Weight:  math.NewInt(502351235),
					},
					{
						GaugeId: 223525,
						Weight:  math.NewInt(53454350),
					},
				},
			},
			hasInitialDistr: true,
			initialVote: sponsorshiptypes.MsgVote{
				Voter: addrs[0].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(70)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(30)},
				},
			},
			hasIntermediateDistr: true,
			intermediateVote: sponsorshiptypes.MsgVote{
				Voter: addrs[1].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: sponsorshiptypes.DYM.MulRaw(10)},
					{GaugeId: 2, Weight: sponsorshiptypes.DYM.MulRaw(90)},
				},
			},
			fillEpochs: true,
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Cast an initial vote
			if tc.hasInitialDistr {
				suite.Vote(tc.initialVote, sponsorshiptypes.DYM)
			}

			// Create a stream
			sID, s := suite.CreateSponsoredStream(tc.stream.distrInfo, tc.stream.coins, time.Now().Add(-time.Minute), "day", tc.stream.numOfEpochs)

			// Check that the stream distr matches the current sponsorship distr
			actualDistr, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
			suite.Require().NoError(err)
			suite.Require().Equal(s, actualDistr)
			initialDistr := suite.Distribution()
			initialDistrInfo := types.DistrInfoFromDistribution(initialDistr)
			suite.Require().Equal(initialDistrInfo.TotalWeight, actualDistr.DistributeTo.TotalWeight)
			suite.Require().ElementsMatch(initialDistrInfo.Records, actualDistr.DistributeTo.Records)

			// Cast an intermediate vote
			if tc.hasIntermediateDistr {
				suite.Vote(tc.intermediateVote, sponsorshiptypes.DYM)
			}

			// Distribute
			// First, simulate the epoch start. This moves gauges from upcoming to active and
			// updates corresponding streams parameters.
			err = suite.App.StreamerKeeper.BeforeEpochStart(suite.Ctx, "day")
			suite.Require().NoError(err)

			// Then, simulate the epoch end. This triggers the distribution of the rewards.
			_, err = suite.App.StreamerKeeper.AfterEpochEnd(suite.Ctx, "day")
			suite.Require().NoError(err)

			// Check that the stream distr matches the current sponsorship distr
			actualDistr, err = suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
			suite.Require().NoError(err)
			intermediateDistr := suite.Distribution()
			intermediateDistrInfo := types.DistrInfoFromDistribution(intermediateDistr)
			suite.Require().Equal(intermediateDistrInfo.TotalWeight, actualDistr.DistributeTo.TotalWeight)
			suite.Require().ElementsMatch(intermediateDistrInfo.Records, actualDistr.DistributeTo.Records)

			// Check the state
			actual, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.fillEpochs, actual.FilledEpochs > 0)

			// Calculate expected rewards. The result is based on the merged initial and intermediate distributions.
			expectedDistr := types.DistrInfoFromDistribution(initialDistr.Merge(intermediateDistr))
			gaugesExpectedRewards := make(map[uint64]sdk.Coins)
			for _, coin := range tc.stream.coins {
				epochAmt := coin.Amount.Quo(math.NewIntFromUint64(tc.stream.numOfEpochs))
				if !epochAmt.IsPositive() {
					continue
				}

				for _, record := range expectedDistr.Records {
					expectedAmtFromStream := epochAmt.Mul(record.Weight).Quo(expectedDistr.TotalWeight)
					expectedCoins := sdk.NewCoin(coin.Denom, expectedAmtFromStream)
					gaugesExpectedRewards[record.GaugeId] = gaugesExpectedRewards[record.GaugeId].Add(expectedCoins)
				}
			}

			// Check expected rewards against actual rewards received
			gauges := suite.App.IncentivesKeeper.GetGauges(suite.Ctx)
			for _, gauge := range gauges {
				suite.Require().ElementsMatch(gaugesExpectedRewards[gauge.Id], gauge.Coins)
			}
		})
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	var err error
	suite.SetupTest()
	err = suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)

	// check that the sum of coins yet to be distributed is nil
	coins := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins{})

	// setup a stream
	streamCoins := sdk.Coins{sdk.NewInt64Coin("stake", 600000)}
	_, _ = suite.CreateDefaultStream(streamCoins)

	// check that the sum of coins yet to be distributed is equal to the newly created streamCoins
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins)

	// create a new stream
	// check that the sum of coins yet to be distributed is equal to the stream1 and stream2 coins combined

	streamCoins2 := sdk.Coins{sdk.NewInt64Coin("udym", 300000)}
	_, _ = suite.CreateDefaultStream(streamCoins2)

	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, streamCoins.Add(streamCoins2...))

	// move all created streams from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(time.Now())
	streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)

	// distribute coins to stakers
	distrCoins := suite.DistributeAllRewards(streams)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 20000), sdk.NewInt64Coin("udym", 10000)}, distrCoins)

	// check stream changes after distribution
	coins = suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().ElementsMatch(coins, streamCoins.Add(streamCoins2...).Sub(distrCoins...))
}
