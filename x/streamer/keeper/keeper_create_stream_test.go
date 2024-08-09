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
	suite.SetupTest()

	currModuleBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, authtypes.NewModuleAddress(types.ModuleName))
	suite.Require().Equal(len(currModuleBalance), 2)
	coins1 := sdk.NewCoins(currModuleBalance[0])
	coins2 := sdk.NewCoins(currModuleBalance[1])

	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins1, defaultDistrInfo, time.Time{}, "day", 30, NonSponsored)
	suite.Require().NoError(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins2, defaultDistrInfo, time.Now().Add(10*time.Minute), "day", 30, NonSponsored)
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
		suite.SetupTest()
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
	}{
		{
			name:              "empty initial distr",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDistrInfo,
			epochIdentifier:   "day",
			initialVote:       sponsorshiptypes.MsgVote{},
			numEpochsPaidOver: 30,
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
		},
	}

	for _, tc := range tests {
		suite.SetupTest()
		sID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, tc.coins, tc.distrTo, time.Time{}, tc.epochIdentifier, tc.numEpochsPaidOver, Sponsored)
		suite.Require().NoError(err, tc.name)

		// Check that the stream distr matches the current sponsorship distr
		actualDistr, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, sID)
		suite.Require().NoError(err)
		initialDistr := suite.Distribution()
		initialDistrInfo := types.DistrInfoFromDistribution(initialDistr)
		suite.Require().Equal(initialDistrInfo.TotalWeight, actualDistr.DistributeTo.TotalWeight)
		suite.Require().ElementsMatch(initialDistrInfo.Records, actualDistr.DistributeTo.Records)
	}
}
