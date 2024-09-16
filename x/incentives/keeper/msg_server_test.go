package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name: "user tries to create a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(40))), // 40 >= 20 (adym) + 10 (creation fee) + 10 (for every denom) = 40
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(5)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(1)))),
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(10)))),
			expectErr:            true,
		},
		{
			name: "user tries to create a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(39))), // 39 < 20 (adym) + 10 (creation fee) + 10 (for every denom) = 40
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			expectErr: true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(30)))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

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
			_, err := msgServer.CreateGauge(sdk.WrapSDKContext(ctx), msg)

			if tc.expectErr {
				suite.Require().Error(err, "test: %v", tc.name)
			} else {
				suite.Require().NoError(err, "test: %v", tc.name)
			}

			balanceAmount := bankKeeper.GetAllBalances(ctx, testAccountAddress)

			if tc.expectErr {
				suite.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
			} else {
				// Fee = CreateGaugeBaseFee + AddDenomFee * NumDenoms
				params := suite.querier.GetParams(suite.Ctx)
				feeRaw := params.CreateGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition))))
				fee := sdk.NewCoins(sdk.NewCoin("stake", feeRaw))

				accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
				finalAccountBalance := accountBalance.Sub(fee...)
				suite.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
			}
		})
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
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(35)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
		},
		{
			name: "user adds to a non-perpetual gauge including many denoms",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(31))), // 31 >= 20 (adym) + 10 (denoms) + 1 (initial denom) = 31
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
		},
		{
			name:                 "module account creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(10)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20)))),
			expectErr:            true,
		},
		{
			name: "user tries to add to a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(30))), // 30 < 20 (adym) + 10 (denoms) + 1 (initial denom) = 31
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(sdk.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(sdk.NewInt(20))),
			),
			expectErr: true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(sdk.NewInt(10)))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "stake")
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
			gaugeID, gauge, _, _ := suite.SetupNewGauge(true, coins)
			if tc.nonexistentGauge {
				gaugeID = incentivesKeeper.GetLastGaugeID(ctx) + 1
			}
			msg := &types.MsgAddToGauge{
				Owner:   testAccountAddress.String(),
				GaugeId: gaugeID,
				Rewards: tc.gaugeAddition,
			}

			params := suite.querier.GetParams(suite.Ctx)
			feeRaw := params.AddToGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition) + len(gauge.Coins))))
			suite.T().Log(feeRaw, params.AddToGaugeBaseFee, params.AddDenomFee)

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
				// Fee = AddToGaugeBaseFee + AddDenomFee * (NumAddedDenoms + NumGaugeDenoms)
				params := suite.querier.GetParams(suite.Ctx)
				feeRaw := params.AddToGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition) + len(gauge.Coins))))
				fee := sdk.NewCoins(sdk.NewCoin("stake", feeRaw))

				accountBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
				finalAccountBalance := accountBalance.Sub(fee...)
				suite.Require().Equal(finalAccountBalance.String(), bal.String(), "test: %v", tc.name)
			}
		})
	}
}
