package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestCreateGauge_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		expectedEndBalance   sdk.Coins
		isPerpetual          bool
		isModuleAccount      bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(5)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(1)))),
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(10)))),
			expectErr:            true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(20)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(30)))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "adym")
		suite.Require().NoError(err)

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		ctx := suite.Ctx
		bankKeeper := suite.App.BankKeeper
		accountKeeper := suite.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(suite.App.IncentivesKeeper)

		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(ctx, modAcc)
		}

		suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
		distrTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         defaultLPDenom,
			Duration:      defaultLockDuration,
		}

		msg := &types.MsgCreateGauge{
			IsPerpetual:       tc.isPerpetual,
			Owner:             testAccountAddress.String(),
			DistributeTo:      distrTo,
			Coins:             tc.gaugeAddition,
			StartTime:         time.Now(),
			NumEpochsPaidOver: 1,
		}
		// System under test.
		_, err = msgServer.CreateGauge(sdk.WrapSDKContext(ctx), msg)

		if tc.expectErr {
			suite.Require().Error(err, "test: %v", tc.name)
		} else {
			suite.Require().NoError(err, "test: %v", tc.name)
		}

		balanceAmount := bankKeeper.GetAllBalances(ctx, testAccountAddress)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
		} else {
			fee := sdk.NewCoins(sdk.NewCoin("adym", types.CreateGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
			finalAccountBalance := accountBalance.Sub(fee...)
			suite.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestAddToGauge_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		nonexistentGauge     bool
		isPerpetual          bool
		isModuleAccount      bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(35)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(20)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("adym", types.DYM.Mul(sdk.NewInt(20)))),
			expectErr:            false, // no addition fee
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(10)))),
			expectErr:            false, // no addition fee
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "adym")
		suite.Require().NoError(err)

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())
		// testAccountAddress := suite.TestAccs[0]

		ctx := suite.Ctx
		bankKeeper := suite.App.BankKeeper
		incentivesKeeper := suite.App.IncentivesKeeper
		accountKeeper := suite.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(incentivesKeeper)

		// suite.FundAcc(testAccountAddress, testutil.DefaultAcctFunds)
		suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

		if tc.isModuleAccount {
			modAcc := authtypes.NewModuleAccount(authtypes.NewBaseAccount(testAccountAddress, testAccountPubkey, 1, 0),
				"module",
				"permission",
			)
			accountKeeper.SetModuleAccount(ctx, modAcc)
		}

		// System under test.
		coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(500000000)))
		gaugeID, _, _, _ := suite.SetupNewGauge(true, coins)
		if tc.nonexistentGauge {
			gaugeID = incentivesKeeper.GetLastGaugeID(ctx) + 1
		}
		msg := &types.MsgAddToGauge{
			Owner:   testAccountAddress.String(),
			GaugeId: gaugeID,
			Rewards: tc.gaugeAddition,
		}

		_, err = msgServer.AddToGauge(sdk.WrapSDKContext(ctx), msg)

		if tc.expectErr {
			suite.Require().Error(err, "test: %v", tc.name)
		} else {
			suite.Require().NoError(err, "test: %v", tc.name)
		}

		bal := bankKeeper.GetAllBalances(ctx, testAccountAddress)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), bal.String(), "test: %v", tc.name)
		} else {
			fee := sdk.NewCoins(sdk.NewCoin("adym", types.AddToGaugeFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
			finalAccountBalance := accountBalance.Sub(fee...)
			suite.Require().Equal(finalAccountBalance.String(), bal.String(), "test: %v", tc.name)
		}
	}
}
