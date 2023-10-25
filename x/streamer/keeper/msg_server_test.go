package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	keeper "github.com/dymensionxyz/dymension/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/x/streamer/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestCreateStream_Fee() {
	tests := []struct {
		name                 string
		accountBalanceToFund sdk.Coins
		streamAddition       sdk.Coins
		expectedEndBalance   sdk.Coins
		isPerpetual          bool
		isModuleAccount      bool
		expectErr            bool
	}{
		{
			name:                 "user creates a non-perpetual stream and fills stream with all remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(60000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
		},
		{
			name:                 "user creates a non-perpetual stream and fills stream with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(70000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
		},
		{
			name:                 "user with multiple denoms creates a non-perpetual stream and fills stream with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
		},
		{
			name:                 "module account creates a perpetual stream and fills stream with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
			isPerpetual:          true,
			isModuleAccount:      true,
		},
		{
			name:                 "user with multiple denoms creates a perpetual stream and fills stream with some remaining tokens",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(70000000)), sdk.NewCoin("foo", sdk.NewInt(70000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
			isPerpetual:          true,
		},
		{
			name:                 "user tries to create a non-perpetual stream but does not have enough funds to pay for the create stream fee",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(40000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(10000000))),
			expectErr:            true,
		},
		{
			name:                 "user tries to create a non-perpetual stream but does not have the correct fee denom",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(60000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(10000000))),
			expectErr:            true,
		},
		{
			name:                 "one user tries to create a stream, has enough funds to pay for the create stream fee but not enough to fill the stream",
			accountBalanceToFund: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(60000000))),
			streamAddition:       sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(30000000))),
			expectErr:            true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		testAccountPubkey := secp256k1.GenPrivKeyFromSecret([]byte("acc")).PubKey()
		testAccountAddress := sdk.AccAddress(testAccountPubkey.Address())

		ctx := suite.Ctx
		bankKeeper := suite.App.BankKeeper
		accountKeeper := suite.App.AccountKeeper
		msgServer := keeper.NewMsgServerImpl(suite.App.StreamerKeeper)

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

		msg := &types.MsgCreateStream{
			IsPerpetual:       tc.isPerpetual,
			Owner:             testAccountAddress.String(),
			DistributeTo:      distrTo,
			Coins:             tc.streamAddition,
			StartTime:         time.Now(),
			NumEpochsPaidOver: 1,
		}
		// System under test.
		_, err := msgServer.CreateStream(sdk.WrapSDKContext(ctx), msg)

		if tc.expectErr {
			suite.Require().Error(err)
		} else {
			suite.Require().NoError(err)
		}

		balanceAmount := bankKeeper.GetAllBalances(ctx, testAccountAddress)

		if tc.expectErr {
			suite.Require().Equal(tc.accountBalanceToFund.String(), balanceAmount.String(), "test: %v", tc.name)
		} else {
			fee := sdk.NewCoins(sdk.NewCoin("udym", types.CreateStreamFee))
			accountBalance := tc.accountBalanceToFund.Sub(tc.streamAddition...)
			finalAccountBalance := accountBalance.Sub(fee...)
			suite.Require().Equal(finalAccountBalance.String(), balanceAmount.String(), "test: %v", tc.name)
		}
	}
}
