package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

// TestNonExistentDenomStreamCreation tests error handling for creating a stream with an invalid denom.
func (suite *KeeperTestSuite) TestNonExistentDenomStreamCreation() {
	suite.SetupTest()

	//udym exists
	coins := sdk.Coins{sdk.NewInt64Coin("udym", 1000)}
	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().NoError(err)

	//udym and stake exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("stake", 10000)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().NoError(err)

	//udym2 doesn't exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym", 1000), sdk.NewInt64Coin("udym2", 1000)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().Error(err)

	//udym2 doesn't exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym2", 10000)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().Error(err)
}

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
		insufficent := !tc.coinsToSpend.IsAllLTE(tc.balance)
		if insufficent != tc.expectErr {
			t.Errorf("%d, expected error: %v, got: %v", i, tc.expectErr, insufficent)
		}
	}
}

func (suite *KeeperTestSuite) TestCreateStream_CoinsSpendable() {
	suite.SetupTest()

	currModuleBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, authtypes.NewModuleAddress(types.ModuleName))
	suite.Require().Equal(len(currModuleBalance), 2)
	coins1 := sdk.NewCoins(currModuleBalance[0])
	coins2 := sdk.NewCoins(currModuleBalance[1])

	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins1, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().NoError(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins2, defaultDestAddr, time.Now().Add(10*time.Minute), "day", 30)
	suite.Require().NoError(err)

	//Check that all tokens are alloceted for distribution
	toDistribute := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(currModuleBalance, toDistribute)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, sdk.Coins{sdk.NewInt64Coin("udym", 100)}, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().Error(err)

	//mint more tokens to the streamer account
	mintCoins := sdk.NewCoins(sdk.NewInt64Coin("udym", 1000000))
	suite.FundModuleAcc(types.ModuleName, mintCoins)

	newToDistribute := suite.App.StreamerKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(toDistribute, newToDistribute)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, mintCoins.Add(mintCoins...), defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().Error(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, sdk.Coins{sdk.NewInt64Coin("udym", 100)}, defaultDestAddr, time.Time{}, "day", 30)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestCreateStream() {
	tests := []struct {
		name              string
		coins             sdk.Coins
		distrTo           sdk.AccAddress
		epochIdentifier   string
		numEpochsPaidOver uint64
		expectErr         bool
	}{
		{
			name:              "happy flow",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDestAddr,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         false,
		},
		{
			name:              "multiple coins",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 100000), sdk.NewInt64Coin("stake", 100000)},
			distrTo:           defaultDestAddr,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         false,
		},
		{
			name:              "non existing denom",
			coins:             sdk.Coins{sdk.NewInt64Coin("udasdas", 10)},
			distrTo:           defaultDestAddr,
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "bad destination addr",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           sdk.AccAddress("dasdasdasdasda"),
			epochIdentifier:   "day",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "bad epoch identifier",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDestAddr,
			epochIdentifier:   "thththt",
			numEpochsPaidOver: 30,
			expectErr:         true,
		},
		{
			name:              "bad num of epochs",
			coins:             sdk.Coins{sdk.NewInt64Coin("udym", 10)},
			distrTo:           defaultDestAddr,
			epochIdentifier:   "day",
			numEpochsPaidOver: 0,
			expectErr:         true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()
		_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, tc.coins, sdk.AccAddress(tc.distrTo), time.Time{}, tc.epochIdentifier, tc.numEpochsPaidOver)
		if tc.expectErr {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}
