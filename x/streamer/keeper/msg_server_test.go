package keeper_test

import (
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
	coins := sdk.Coins{sdk.NewInt64Coin("udym", 10)}
	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 1)
	suite.Require().NoError(err)

	//udym and stake exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym", 10), sdk.NewInt64Coin("stake", 10)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 1)
	suite.Require().NoError(err)

	//udym2 doesn't exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym", 10), sdk.NewInt64Coin("udym2", 10)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 1)
	suite.Require().Error(err)

	//udym2 doesn't exist
	coins = sdk.Coins{sdk.NewInt64Coin("udym2", 10)}
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, defaultDestAddr, time.Time{}, "day", 1)
	suite.Require().Error(err)
}

// TODO: test stream coins are spendable.
func (suite *KeeperTestSuite) TestCreateStream_CoinsSpendable() {
	suite.SetupTest()

	currModuleBalance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, authtypes.NewModuleAddress(types.ModuleName))

	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, sdk.NewCoins(currModuleBalance[0]), defaultDestAddr, time.Time{}, "day", 1)
	suite.Require().NoError(err)

	// TODO:  check after funding the module again
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
