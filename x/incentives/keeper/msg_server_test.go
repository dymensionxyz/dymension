package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/x/txfees"

	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestCreateGauge() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		isPerpetual          bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(6)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(1)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(7)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(1)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(7))), sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(7)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(1)))),
		},
		{
			name: "user tries to create a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees - 1",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(40))), // 40 >= 2 (adym) + 1 (creation fee) + 1 (for every denom) = 4
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(2))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(2))),
			),
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(7))), sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(1)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(5))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1))),
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(6)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(1)))),
			expectErr:            true,
		},
		{
			name: "user tries to create a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees - 2",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(3))), // 3 < 2 (adym) + 1 (creation fee) + 1 (for every denom) = 4
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(2))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(2))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(2))),
			),
			expectErr: true,
		},
		{
			name:                 "one user tries to create a gauge, has enough funds to pay for the create gauge fee but not enough to fill the gauge",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(2)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(3)))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "stake")
			suite.Require().NoError(err)

			ctx := suite.Ctx
			bankKeeper := suite.App.BankKeeper
			accountKeeper := suite.App.AccountKeeper
			msgServer := keeper.NewMsgServerImpl(suite.App.IncentivesKeeper)
			testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
			testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())
			suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

			suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
			distrTo := lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         defaultLPDenom,
				Duration:      defaultLockDuration,
			}

			msg := &types.MsgCreateGauge{
				IsPerpetual: tc.isPerpetual,
				Owner:       testAccountAddress.String(),
				DistributeTo: &types.MsgCreateGauge_Asset{
					Asset: &distrTo,
				},
				Coins:             tc.gaugeAddition,
				StartTime:         time.Now(),
				NumEpochsPaidOver: 1,
			}
			txfeesBalanceBefore := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(txfees.ModuleName), "stake")

			// System under test.
			_, err = msgServer.CreateGauge(ctx, msg)

			if tc.expectErr {
				suite.Require().Error(err, "test: %v", tc.name)
			} else {
				suite.Require().NoError(err, "test: %v", tc.name)

				// Fee = CreateGaugeBaseFee + AddDenomFee * NumDenoms
				params := suite.querier.GetParams(suite.Ctx)
				feeRaw := params.CreateGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition))))
				fee := sdk.NewCoins(sdk.NewCoin("stake", feeRaw))

				balanceAmount := bankKeeper.GetAllBalances(ctx, testAccountAddress)

				expectedBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
				expectedBalance = expectedBalance.Sub(fee...)
				suite.Require().Equal(expectedBalance.String(), balanceAmount.String(), "test: %v", tc.name)

				// test fee charged to txfees module account and burned
				txfeesBalanceAfter := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(txfees.ModuleName), "stake")
				suite.Require().Equal(txfeesBalanceBefore.Amount, txfeesBalanceAfter.Amount, "test: %v", tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestAddToGauge() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		gaugeAddition        sdk.Coins
		isPerpetual          bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(35)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(10)))),
		},
		{
			name:                 "user creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(10)))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(10)))),
		},
		{
			name: "user adds to a non-perpetual gauge including many denoms",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(31))), // 31 >= 20 (adym) + 10 (denoms) + 1 (initial denom) = 31
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(20))),
			),
		},
		{
			name:                 "user with multiple denoms creates a perpetual gauge and fills gauge with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(70))), sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(70)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(10)))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have enough funds to pay for the create gauge fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(20)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(20)))),
			expectErr:            true,
		},
		{
			name: "user tries to add to a non-perpetual gauge but includes too many denoms so does not have enough funds to pay fees",
			accountBalanceToFund: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(30))), // 30 < 20 (adym) + 10 (denoms) + 1 (initial denom) = 31
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(20))),
			),
			gaugeAddition: sdk.NewCoins(
				sdk.NewCoin("stake", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("osmo", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("atom", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("abcd", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("efgh", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("igkl", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("mnop", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("qrst", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("uvwx", types.DYM.Mul(math.NewInt(20))),
				sdk.NewCoin("yzzz", types.DYM.Mul(math.NewInt(20))),
			),
			expectErr: true,
		},
		{
			name:                 "user tries to add to a non-perpetual gauge but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(60)))),
			gaugeAddition:        sdk.NewCoins(sdk.NewCoin("foo", types.DYM.Mul(math.NewInt(10)))),
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

			ctx := suite.Ctx
			bankKeeper := suite.App.BankKeeper
			incentivesKeeper := suite.App.IncentivesKeeper
			accountKeeper := suite.App.AccountKeeper
			msgServer := keeper.NewMsgServerImpl(incentivesKeeper)

			suite.FundAcc(testAccountAddress, tc.accountBalanceToFund)

			// System under test.
			coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(500000000)))
			gaugeID, gauge, _, _ := suite.SetupNewGauge(true, coins)
			msg := &types.MsgAddToGauge{
				Owner:   testAccountAddress.String(),
				GaugeId: gaugeID,
				Rewards: tc.gaugeAddition,
			}

			params := suite.querier.GetParams(suite.Ctx)
			feeRaw := params.AddToGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition) + len(gauge.Coins))))
			suite.T().Log(feeRaw, params.AddToGaugeBaseFee, params.AddDenomFee)

			txfeesBalanceBefore := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(txfees.ModuleName), "stake")

			_, err = msgServer.AddToGauge(sdk.WrapSDKContext(ctx), msg)

			if tc.expectErr {
				suite.Require().Error(err, "test: %v", tc.name)
			} else {
				suite.Require().NoError(err, "test: %v", tc.name)

				// Fee = AddToGaugeBaseFee + AddDenomFee * (NumAddedDenoms + NumGaugeDenoms)
				params := suite.querier.GetParams(suite.Ctx)
				feeRaw := params.AddToGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(tc.gaugeAddition) + len(gauge.Coins))))
				fee := sdk.NewCoins(sdk.NewCoin("stake", feeRaw))

				bal := bankKeeper.GetAllBalances(ctx, testAccountAddress)
				expectedBalance := tc.accountBalanceToFund.Sub(tc.gaugeAddition...)
				expectedBalance = expectedBalance.Sub(fee...)
				suite.Require().Equal(expectedBalance.String(), bal.String(), "test: %v", tc.name)

				// test fee charged to txfees module account and burned
				txfeesBalanceAfter := bankKeeper.GetBalance(ctx, accountKeeper.GetModuleAddress(txfees.ModuleName), "stake")
				suite.Require().Equal(txfeesBalanceBefore.Amount, txfeesBalanceAfter.Amount, "test: %v", tc.name)
			}
		})
	}
}
