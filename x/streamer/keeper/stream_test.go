package keeper_test

import (
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

// TestInvalidDurationStreamCreationValidation tests error handling for creating a stream with an invalid duration.
func (suite *KeeperTestSuite) TestInvalidDurationStreamCreationValidation() {
	suite.SetupTest()

	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestNonExistentDenomStreamCreation tests error handling for creating a stream with an invalid denom.
func (suite *KeeperTestSuite) TestNonExistentDenomStreamCreation() {
	suite.SetupTest()

	addrNoSupply := sdk.AccAddress([]byte("Stream_Creation_Addr_"))
	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration,
	}
	_, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, false, addrNoSupply, defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	_, err = suite.App.StreamerKeeper.CreateStream(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestStreamOperations tests perpetual and non-perpetual stream distribution logic using the streams by denom keeper.
func (suite *KeeperTestSuite) TestStreamOperations() {
	testCases := []struct {
		isPerpetual bool
		numLocks    int
	}{
		{
			isPerpetual: true,
			numLocks:    1,
		},
		{
			isPerpetual: false,
			numLocks:    1,
		},
		{
			isPerpetual: true,
			numLocks:    2,
		},
		{
			isPerpetual: false,
			numLocks:    2,
		},
	}
	for _, tc := range testCases {
		// test for module get streams
		suite.SetupTest()

		// initial module streams check
		streams := suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
		suite.Require().Len(streams, 0)
		streamIdsByDenom := suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(streamIdsByDenom, 0)

		// setup lock and stream
		lockOwners := suite.SetupManyLocks(tc.numLocks, defaultLiquidTokens, defaultLPTokens, time.Second)
		streamID, _, coins, startTime := suite.SetupNewStream(tc.isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 12)})
		// evenly distributed per lock
		expectedCoinsPerLock := sdk.Coins{sdk.NewInt64Coin("stake", 12/int64(tc.numLocks))}
		// set expected epochs
		var expectedNumEpochsPaidOver int
		if tc.isPerpetual {
			expectedNumEpochsPaidOver = 1
		} else {
			expectedNumEpochsPaidOver = 2
		}

		// check streams
		streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
		suite.Require().Len(streams, 1)
		expectedStream := types.Stream{
			Id:          streamID,
			IsPerpetual: tc.isPerpetual,
			DistributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "lptoken",
				Duration:      time.Second,
			},
			Coins:             coins,
			NumEpochsPaidOver: uint64(expectedNumEpochsPaidOver),
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins{},
			StartTime:         startTime,
		}
		suite.Require().Equal(expectedStream.String(), streams[0].String())

		// check stream ids by denom
		streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(streamIdsByDenom, 1)
		suite.Require().Equal(streamID, streamIdsByDenom[0])

		// check rewards estimation
		rewardsEst := suite.App.StreamerKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
		suite.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

		// check streams
		streams = suite.App.StreamerKeeper.GetNotFinishedStreams(suite.Ctx)
		suite.Require().Len(streams, 1)
		suite.Require().Equal(expectedStream.String(), streams[0].String())

		// check upcoming streams
		streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
		suite.Require().Len(streams, 1)

		// start distribution
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
		suite.Require().NoError(err)
		err = suite.App.StreamerKeeper.MoveUpcomingStreamToActiveStream(suite.Ctx, *stream)
		suite.Require().NoError(err)

		// check active streams
		streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
		suite.Require().Len(streams, 1)

		// check upcoming streams
		streams = suite.App.StreamerKeeper.GetUpcomingStreams(suite.Ctx)
		suite.Require().Len(streams, 0)

		// check stream ids by denom
		streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(streamIdsByDenom, 1)

		// check stream ids by other denom
		streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lpt")
		suite.Require().Len(streamIdsByDenom, 0)

		// distribute coins to stakers
		distrCoins, err := suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
		suite.Require().NoError(err)
		// We hardcoded 12 "stake" tokens when initializing stream
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

		if tc.isPerpetual {
			// distributing twice without adding more for perpetual stream
			stream, err = suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
			suite.Require().NoError(err)
			suite.Require().True(distrCoins.Empty())

			// add to stream
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToStream(addCoins, streamID)

			// distributing twice with adding more for perpetual stream
			stream, err = suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.StreamerKeeper.Distribute(suite.Ctx, []types.Stream{*stream})
			suite.Require().NoError(err)
			suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 200)}, distrCoins)
		} else {
			// add to stream
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToStream(addCoins, streamID)
		}

		// check active streams
		streams = suite.App.StreamerKeeper.GetActiveStreams(suite.Ctx)
		suite.Require().Len(streams, 1)

		// check stream ids by denom
		streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(streamIdsByDenom, 1)

		// finish distribution for non perpetual stream
		if !tc.isPerpetual {
			err = suite.App.StreamerKeeper.MoveActiveStreamToFinishedStream(suite.Ctx, *stream)
			suite.Require().NoError(err)
		}

		// check non-perpetual streams (finished + rewards estimate empty)
		if !tc.isPerpetual {

			// check finished streams
			streams = suite.App.StreamerKeeper.GetFinishedStreams(suite.Ctx)
			suite.Require().Len(streams, 1)

			// check stream by ID
			stream, err = suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
			suite.Require().NoError(err)
			suite.Require().NotNil(stream)
			suite.Require().Equal(streams[0], *stream)

			// check invalid stream ID
			_, err = suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID+1000)
			suite.Require().Error(err)
			rewardsEst = suite.App.StreamerKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins{}, rewardsEst)

			// check stream ids by denom
			streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(streamIdsByDenom, 0)
		} else { // check perpetual streams (not finished + rewards estimate empty)

			// check finished streams
			streams = suite.App.StreamerKeeper.GetFinishedStreams(suite.Ctx)
			suite.Require().Len(streams, 0)

			// check rewards estimation
			rewardsEst = suite.App.StreamerKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins(nil), rewardsEst)

			// check stream ids by denom
			streamIdsByDenom = suite.App.StreamerKeeper.GetAllStreamIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(streamIdsByDenom, 1)
		}
	}
}

func (suite *KeeperTestSuite) TestChargeFeeIfSufficientFeeDenomBalance() {
	const baseFee = int64(100)

	testcases := map[string]struct {
		accountBalanceToFund sdk.Coin
		feeToCharge          int64
		streamCoins          sdk.Coins

		expectError bool
	}{
		"fee + base denom stream coin == acount balance, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(baseFee/2))),
		},
		"fee + base denom stream coin < acount balance, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 - 1,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(baseFee/2))),
		},
		"fee + base denom stream coin > acount balance, error": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 + 1,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(baseFee/2))),
			expectError:          true,
		},
		"fee + base denom stream coin < acount balance, custom values, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(11793193112)),
			feeToCharge:          55,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(328812))),
		},
		"account funded with coins other than base denom, error": {
			accountBalanceToFund: sdk.NewCoin("usdc", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(baseFee/2))),
			expectError:          true,
		},
		"fee == account balance, no stream coins, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
		},
		"stream coins == account balance, no fee, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			streamCoins:          sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(baseFee))),
		},
		"fee == account balance, stream coins in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2))),
		},
		"fee + stream coins == account balance, multiple stream coins, one in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin("udym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			streamCoins:          sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2)), sdk.NewCoin("udym", sdk.NewInt(baseFee/2))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()

			testAccount := suite.TestAccs[0]

			ctx := suite.Ctx
			StreamerKeepers := suite.App.StreamerKeeper
			bankKeeper := suite.App.BankKeeper

			// Pre-fund account.
			// suite.FundAcc(testAccount, testutil.DefaultAcctFunds)
			suite.FundAcc(testAccount, sdk.NewCoins(tc.accountBalanceToFund))

			oldBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, "udym").Amount

			// System under test.
			err := StreamerKeepers.ChargeFeeIfSufficientFeeDenomBalance(ctx, testAccount, sdk.NewInt(tc.feeToCharge), tc.streamCoins)

			// Assertions.
			newBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, "udym").Amount
			if tc.expectError {
				suite.Require().Error(err)

				// check account balance unchanged
				suite.Require().Equal(oldBalanceAmount, newBalanceAmount)
			} else {
				suite.Require().NoError(err)

				// check account balance changed.
				expectedNewBalanceAmount := oldBalanceAmount.Sub(sdk.NewInt(tc.feeToCharge))
				suite.Require().Equal(expectedNewBalanceAmount.String(), newBalanceAmount.String())
			}
		})
	}
}
