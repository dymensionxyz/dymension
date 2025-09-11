package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

var _ = suite.TestingSuite(nil)

func TestSpendable(t *testing.T) {
	tests := []struct {
		balance      sdk.Coins
		coinsToSpend sdk.Coins
		expectErr    bool
	}{
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000)},
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1000)},
			expectErr:    false,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000)},
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1001)},
			expectErr:    true,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1000)},
			expectErr:    false,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			expectErr:    false,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1001)}.Sort(),
			expectErr:    true,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1001), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			expectErr:    true,
		},
		{
			balance:      sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 1000)}.Sort(),
			coinsToSpend: sdk.Coins{sdk.NewInt64Coin("udym", 1001), sdk.NewInt64Coin("stake", 1001)}.Sort(),
			expectErr:    true,
		},
	}
	for i, tc := range tests {
		insufficient := !tc.coinsToSpend.IsAllLTE(tc.balance)
		if insufficient != tc.expectErr {
			t.Errorf("%d, expected error: %v, got: %v", i, tc.expectErr, insufficient)
		}
	}
}

func (suite *KeeperTestSuite) TestCreateStream_CoinsSpendable() {
	currModuleBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, authtypes.NewModuleAddress(types.ModuleName))
	suite.Require().Equal(len(currModuleBalance), 3)
	coins1 := sdk.NewCoins(currModuleBalance[0])
	coins2 := sdk.NewCoins(currModuleBalance[1])
	coins3 := sdk.NewCoins(currModuleBalance[2])

	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins1, defaultDistrInfo, time.Time{}, "day", 30, NonSponsored)
	suite.Require().NoError(err)

	_, err = suite.App.StreamerKeeper.CreatePumpStream(suite.Ctx, types.CreateStreamGeneric{
		Coins:             coins2,
		StartTime:         time.Now().Add(10 * time.Minute),
		EpochIdentifier:   "day",
		NumEpochsPaidOver: 30,
	}, 1, 1, types.PumpTargetRollapps(1))
	suite.Require().NoError(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins3, defaultDistrInfo, time.Now().Add(10*time.Minute), "day", 30, Sponsored)
	suite.Require().NoError(err)

	// Check that all tokens are alloceted for distribution
	toDistribute := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(currModuleBalance, toDistribute)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, sdk.Coins{sdk.NewInt64Coin("udym", 100)}, defaultDistrInfo, time.Time{}, "day", 30, NonSponsored)
	suite.Require().Error(err)

	// mint more tokens to the streamer account
	mintCoins := sdk.NewCoins(sdk.NewInt64Coin("udym", 1000000))
	suite.FundModuleAcc(types.ModuleName, mintCoins)

	newToDistribute := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(toDistribute, newToDistribute)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, mintCoins.Add(mintCoins...), defaultDistrInfo, time.Time{}, "day", 30, NonSponsored)
	suite.Require().Error(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, sdk.Coins{sdk.NewInt64Coin("udym", 100)}, defaultDistrInfo, time.Time{}, "day", 30, NonSponsored)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestCreateStream() {
	tests := []struct {
		name              string
		coins             sdk.Coins
		distrTo           []types.DistrRecord
		epochIdentifier   string
		numEpochsPaidOver uint64
		expectErr         bool
	}{
		{
			name:              "happy flow",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         false,
		},
		{
			name:              "multiple coins",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("stake", 100000)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         false,
		},
		{
			name:              "non existing denom",
			coins:             sdk.Coins{sdk.NewInt64Coin("udasdas", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "multiple tokens - one is non existing denom",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("udasdas", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:  "bad distribution info - negative weight",
			coins: sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo: []types.DistrRecord{
				{
					GaugeId: 1,
					Weight:  math.NewInt(-1),
				},
			},
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:  "bad distribution info - invalid gauge",
			coins: sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo: []types.DistrRecord{
				{
					GaugeId: 0,
					Weight:  math.NewInt(10),
				},
			},
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:  "bad distribution info - zero weight",
			coins: sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo: []types.DistrRecord{
				{
					GaugeId: 2,
					Weight:  math.NewInt(0),
				},
			},
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "bad epoch identifier",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "thththt",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "bad num of epochs",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			numEpochsPaidOver: 0,
			expectErr:         true,
		},
	}

	for _, tc := range tests {
		_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, tc.coins, tc.distrTo, time.Time{}, tc.epochIdentifier, tc.numEpochsPaidOver, NonSponsored)
		if tc.expectErr {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestCreateSponsoredStream() {
	addrs := apptesting.CreateRandomAccounts(3)

	tests := []struct {
		name              string
		coins             sdk.Coins
		distrTo           []types.DistrRecord
		epochIdentifier   string
		initialVote       sponsorshiptypes.MsgVote // the vote that forms the initial distribution
		numEpochsPaidOver uint64
		sponsored         bool
		expectErr         bool
	}{
		{
			name:              "empty initial distr",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			initialVote:       sponsorshiptypes.MsgVote{},
			numEpochsPaidOver: 30,
			sponsored:         true,
			expectErr:         false,
		},
		{
			name:            "non-empty initial distr",
			coins:           sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("stake", 100000)},
			distrTo:         defaultDistrInfo,
			epochIdentifier: "day",
			initialVote: sponsorshiptypes.MsgVote{
				Voter: addrs[0].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: math.NewInt(50)},
					{GaugeId: 2, Weight: math.NewInt(30)},
				},
			},
			numEpochsPaidOver: 30,
			sponsored:         true,
			expectErr:         false,
		},
		{
			name:  "stream distr info doesn't play any role",
			coins: sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("stake", 100000)},
			// Random unrealistic values
			distrTo: []types.DistrRecord{
				{
					GaugeId: 121424,
					Weight:  math.NewInt(502351235),
				},
				{
					GaugeId: 223525,
					Weight:  math.NewInt(53454350),
				},
			},
			epochIdentifier: "day",
			initialVote: sponsorshiptypes.MsgVote{
				Voter: addrs[0].String(),
				Weights: []sponsorshiptypes.GaugeWeight{
					{GaugeId: 1, Weight: math.NewInt(50)},
					{GaugeId: 2, Weight: math.NewInt(30)},
				},
			},
			numEpochsPaidOver: 30,
			sponsored:         true,
			expectErr:         false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			sID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, tc.coins, tc.distrTo, time.Time{}, tc.epochIdentifier, tc.numEpochsPaidOver, tc.sponsored)

			if tc.expectErr {
				suite.Require().Error(err)

				// Verify no stream was actually created by checking that GetStreamByID fails
				streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)
				suite.Require().Empty(streams)
			} else {
				suite.Require().NoError(err)

				// Check that the stream distr matches the current sponsorship distr
				actualDistr, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
				suite.Require().NoError(err)
				initialDistr := suite.Distribution()
				initialDistrInfo := types.DistrInfoFromDistribution(initialDistr)
				suite.Require().Equal(initialDistrInfo.TotalWeight, actualDistr.DistributeTo.TotalWeight)
				suite.Require().ElementsMatch(initialDistrInfo.Records, actualDistr.DistributeTo.Records)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreatePumpStream() {
	tests := []struct {
		name              string
		coins             sdk.Coins
		epochIdentifier   string
		numEpochsPaidOver uint64
		numPumps          uint64
		pumpDistr         types.PumpDistr
		target            types.PumpTarget
		expectErr         bool
	}{
		{
			name:              "happy flow - basic pump stream",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         false,
		},
		{
			name:              "pump stream with multiple coins should work",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)}, // Only DYM allowed for pump streams
			epochIdentifier:   "day",
			numEpochsPaidOver: 5,
			numPumps:          500,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(3),
			expectErr:         false,
		},
		{
			name:              "pump stream with zero NumTopRollapps should fail",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(0),
			expectErr:         true,
		},
		{
			name:              "pump stream with zero NumPumps should fail",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          0,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
		{
			name:              "pump stream with UNSPECIFIED PumpDistr should fail",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNSPECIFIED,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
		{
			name:              "non-udym denom should fail for pump stream",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
		{
			name:              "multiple coins with pump params should fail",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
		{
			name:              "bad epoch identifier with pump params",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "invalid_epoch",
			numEpochsPaidOver: 10,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
		{
			name:              "bad num of epochs with pump params",
			coins:             sdk.Coins{sdk.NewInt64Coin("stake", 100000)},
			epochIdentifier:   "day",
			numEpochsPaidOver: 0,
			numPumps:          1000,
			pumpDistr:         types.PumpDistr_PUMP_DISTR_UNIFORM,
			target:            types.PumpTargetRollapps(2),
			expectErr:         true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			sID, err := suite.App.StreamerKeeper.CreatePumpStream(suite.Ctx, types.CreateStreamGeneric{
			Coins:             tc.coins,
			StartTime:         time.Time{},
			EpochIdentifier:   tc.epochIdentifier,
			NumEpochsPaidOver: tc.numEpochsPaidOver,
		}, tc.numPumps, tc.pumpDistr, tc.target)

			if tc.expectErr {
				suite.Require().Error(err, tc.name)

				// Verify no stream was actually created
				streams := suite.App.StreamerKeeper.GetStreams(suite.Ctx)
				suite.Require().Empty(streams)
			} else {
				suite.Require().NoError(err, tc.name)

				// Verify stream was created successfully
				stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
				suite.Require().NoError(err)

				// Verify it's a pump stream
				suite.Require().True(stream.IsPumpStream())

				// Verify pump params are set correctly
				suite.Require().NotNil(stream.PumpParams)
				suite.Require().Equal(tc.numPumps, stream.PumpParams.NumPumps)
				suite.Require().Equal(tc.pumpDistr, stream.PumpParams.PumpDistr)
				switch t := tc.target.(type) {
				case *types.MsgCreatePumpStream_Rollapps:
					suite.IsType(&types.PumpParams_Rollapps{}, stream.PumpParams.Target)
					actual := stream.PumpParams.Target.(*types.PumpParams_Rollapps)
					suite.Require().Equal(t.Rollapps.NumTopRollapps, actual.Rollapps.NumTopRollapps)
				case *types.MsgCreatePumpStream_Pool:
					suite.IsType(&types.PumpParams_Pool{}, stream.PumpParams.Target)
					actual := stream.PumpParams.Target.(*types.PumpParams_Pool)
					suite.Require().Equal(t.Pool.TokenOut, actual.Pool.TokenOut)
				}

				// Verify epoch budget is correctly calculated
				expectedEpochBudget := tc.coins.QuoInt(math.NewIntFromUint64(tc.numEpochsPaidOver))
				suite.Require().Equal(expectedEpochBudget, stream.EpochCoins)
				suite.Require().Equal(expectedEpochBudget, stream.PumpParams.EpochCoinsLeft)

				// Verify stream is not sponsored
				suite.Require().False(stream.Sponsored)

				// Verify basic stream properties
				suite.Require().Equal(tc.coins, stream.Coins)
				suite.Require().Equal(tc.epochIdentifier, stream.DistrEpochIdentifier)
				suite.Require().Equal(tc.numEpochsPaidOver, stream.NumEpochsPaidOver)
				suite.Require().Equal(uint64(0), stream.FilledEpochs)
				suite.Require().Empty(stream.DistributedCoins)
			}
		})
	}
}
