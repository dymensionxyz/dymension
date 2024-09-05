package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var _ = suite.TestingSuite(nil)

// TestInvalidDurationGaugeCreationValidation tests error handling for creating a gauge with an invalid duration.
func (suite *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	suite.SetupTest()

	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestNonExistentDenomGaugeCreation tests error handling for creating a gauge with an invalid denom.
func (suite *KeeperTestSuite) TestNonExistentDenomGaugeCreation() {
	suite.SetupTest()

	addrNoSupply := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration,
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrNoSupply, defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TestGaugeOperations tests perpetual and non-perpetual gauge distribution logic using the gauges by denom keeper.
func (suite *KeeperTestSuite) TestGaugeOperations() {
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
		// test for module get gauges
		suite.SetupTest()

		// initial module gauges check
		gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)
		gaugeIdsByDenom := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// setup lock and gauge
		_ = suite.SetupManyLocks(tc.numLocks, defaultLiquidTokens, defaultLPTokens, time.Second)
		gaugeID, _, coins, startTime := suite.SetupNewGauge(tc.isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 12)})
		// set expected epochs
		var expectedNumEpochsPaidOver int
		if tc.isPerpetual {
			expectedNumEpochsPaidOver = 1
		} else {
			expectedNumEpochsPaidOver = 2
		}

		// check gauges
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
		expectedGauge := types.Gauge{
			Id:          gaugeID,
			IsPerpetual: tc.isPerpetual,
			DistributeTo: &types.Gauge_Asset{Asset: &lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "lptoken",
				Duration:      time.Second,
			}},
			Coins:             coins,
			NumEpochsPaidOver: uint64(expectedNumEpochsPaidOver),
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins{},
			StartTime:         startTime,
		}
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)
		suite.Require().Equal(gaugeID, gaugeIdsByDenom[0])

		// check gauges
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// start distribution
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
		suite.Require().NoError(err)
		err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		suite.Require().NoError(err)

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// check gauge ids by other denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lpt")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// distribute coins to stakers
		distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
		suite.Require().NoError(err)
		// We hardcoded 12 "stake" tokens when initializing gauge
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

		if tc.isPerpetual {
			// distributing twice without adding more for perpetual gauge
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().True(distrCoins.Empty())

			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToGauge(addCoins, gaugeID)

			// distributing twice with adding more for perpetual gauge
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 200)}, distrCoins)
		} else {
			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToGauge(addCoins, gaugeID)
		}

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// finish distribution for non perpetual gauge
		if !tc.isPerpetual {
			err = suite.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
		}

		// check non-perpetual gauges (finished + rewards estimate empty)
		if !tc.isPerpetual {

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 1)

			// check gauge by ID
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			suite.Require().NotNil(gauge)
			suite.Require().Equal(gauges[0], *gauge)

			// check invalid gauge ID
			_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID+1000)
			suite.Require().Error(err)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 0)
		} else { // check perpetual gauges (not finished + rewards estimate empty)

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 0)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 1)
		}
	}
}

func (suite *KeeperTestSuite) TestChargeFeeIfSufficientFeeDenomBalance() {
	const baseFee = int64(100)

	testcases := map[string]struct {
		accountBalanceToFund sdk.Coin
		feeToCharge          int64
		gaugeCoins           sdk.Coins

		expectError bool
	}{
		"fee + base denom gauge coin == acount balance, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin < acount balance, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 - 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(baseFee/2))),
		},
		"fee + base denom gauge coin > acount balance, error": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee/2 + 1,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(baseFee/2))),
			expectError:          true,
		},
		"fee + base denom gauge coin < acount balance, custom values, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(11793193112)),
			feeToCharge:          55,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(328812))),
		},
		"account funded with coins other than base denom, error": {
			accountBalanceToFund: sdk.NewCoin("usdc", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(baseFee/2))),
			expectError:          true,
		},
		"fee == account balance, no gauge coins, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
		},
		"gauge coins == account balance, no fee, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("adym", sdk.NewInt(baseFee))),
		},
		"fee == account balance, gauge coins in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2))),
		},
		"fee + gauge coins == account balance, multiple gauge coins, one in denom other than base, success": {
			accountBalanceToFund: sdk.NewCoin("adym", sdk.NewInt(baseFee)),
			feeToCharge:          baseFee / 2,
			gaugeCoins:           sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(baseFee*2)), sdk.NewCoin("adym", sdk.NewInt(baseFee/2))),
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.SetupTest()

			err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "adym")
			suite.Require().NoError(err)

			testAccount := apptesting.CreateRandomAccounts(1)[0]
			ctx := suite.Ctx
			incentivesKeepers := suite.App.IncentivesKeeper
			bankKeeper := suite.App.BankKeeper

			// Pre-fund account.
			// suite.FundAcc(testAccount, testutil.DefaultAcctFunds)
			suite.FundAcc(testAccount, sdk.NewCoins(tc.accountBalanceToFund))

			oldBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, "adym").Amount

			// System under test.
			err = incentivesKeepers.ChargeFeeIfSufficientFeeDenomBalance(ctx, testAccount, sdk.NewInt(tc.feeToCharge), tc.gaugeCoins)

			// Assertions.
			newBalanceAmount := bankKeeper.GetBalance(ctx, testAccount, "adym").Amount
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

func (suite *KeeperTestSuite) TestAddToGaugeRewards() {
	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	addr := apptesting.CreateRandomAccounts(1)[0]

	testCases := []struct {
		name               string
		owner              sdk.AccAddress
		coinsToAdd         sdk.Coins
		gaugeCoins         sdk.Coins
		gaugeId            uint64
		minimumGasConsumed uint64

		expectErr bool
	}{
		{
			name:  "valid case: valid gauge",
			owner: addr,
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
			),
			gaugeCoins: sdk.Coins{
				sdk.NewInt64Coin("stake1", 12),
			},
			gaugeId:            1,
			minimumGasConsumed: 3 * params.BaseGasFeeForAddRewardToGauge,
			expectErr:          false,
		},
		{
			name:  "valid case: valid gauge with >4 denoms to add",
			owner: addr,
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
				sdk.NewCoin("mars", sdk.NewInt(88888)),
				sdk.NewCoin("akash", sdk.NewInt(77777)),
				sdk.NewCoin("eth", sdk.NewInt(6666)),
				sdk.NewCoin("usdc", sdk.NewInt(555)),
				sdk.NewCoin("dai", sdk.NewInt(4444)),
				sdk.NewCoin("ust", sdk.NewInt(3333)),
			),
			gaugeCoins: sdk.Coins{
				sdk.NewInt64Coin("stake1", 12),
			},
			gaugeId:            1,
			minimumGasConsumed: 9 * params.BaseGasFeeForAddRewardToGauge,
			expectErr:          false,
		},
		{
			name:  "valid case: valid gauge with >4 initial denoms",
			owner: addr,
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
				sdk.NewCoin("mars", sdk.NewInt(88888)),
				sdk.NewCoin("akash", sdk.NewInt(77777)),
				sdk.NewCoin("eth", sdk.NewInt(6666)),
				sdk.NewCoin("usdc", sdk.NewInt(555)),
				sdk.NewCoin("dai", sdk.NewInt(4444)),
				sdk.NewCoin("ust", sdk.NewInt(3333)),
			),
			gaugeCoins: sdk.Coins{
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
				sdk.NewCoin("mars", sdk.NewInt(88888)),
				sdk.NewCoin("akash", sdk.NewInt(77777)),
				sdk.NewCoin("eth", sdk.NewInt(6666)),
				sdk.NewCoin("usdc", sdk.NewInt(555)),
				sdk.NewCoin("dai", sdk.NewInt(4444)),
				sdk.NewCoin("ust", sdk.NewInt(3333)),
			},
			gaugeId:            1,
			minimumGasConsumed: 16 * params.BaseGasFeeForAddRewardToGauge,
			expectErr:          false,
		},
		{
			name:  "invalid case: gauge Id is not valid",
			owner: addr,
			coinsToAdd: sdk.NewCoins(
				sdk.NewCoin("uosmo", sdk.NewInt(100000)),
				sdk.NewCoin("atom", sdk.NewInt(99999)),
			),
			gaugeCoins: sdk.Coins{
				sdk.NewInt64Coin("stake1", 12),
				sdk.NewInt64Coin("stake2", 12),
				sdk.NewInt64Coin("stake3", 12),
			},
			gaugeId:            0,
			minimumGasConsumed: uint64(0),
			expectErr:          true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			_, _, existingGaugeCoins, _ := suite.SetupNewGauge(true, sdk.NewCoins(tc.gaugeCoins...))

			suite.FundAcc(tc.owner, tc.coinsToAdd)

			existingGasConsumed := suite.Ctx.GasMeter().GasConsumed()

			err := suite.App.IncentivesKeeper.AddToGaugeRewards(suite.Ctx, tc.owner, tc.coinsToAdd, tc.gaugeId)
			if tc.expectErr {
				suite.Require().Error(err)

				// balance shouldn't change in the module
				balance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				suite.Require().Equal(existingGaugeCoins, balance)

			} else {
				suite.Require().NoError(err)

				// Ensure that at least the minimum amount of gas was charged (based on number of additional gauge coins)
				gasConsumed := suite.Ctx.GasMeter().GasConsumed() - existingGasConsumed
				fmt.Println(gasConsumed, tc.minimumGasConsumed)
				suite.Require().True(gasConsumed >= tc.minimumGasConsumed)

				// existing coins gets added to the module when we create gauge and add to gauge
				expectedCoins := existingGaugeCoins.Add(tc.coinsToAdd...)

				// check module account balance, should go up
				balance := suite.App.BankKeeper.GetAllBalances(suite.Ctx, suite.App.AccountKeeper.GetModuleAddress(types.ModuleName))
				suite.Require().Equal(expectedCoins, balance)

				// check gauge coins should go up
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, tc.gaugeId)
				suite.Require().NoError(err)

				suite.Require().Equal(expectedCoins, gauge.Coins)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRapidTestAddToGaugeRewards() {
	rapid.Check(s.T(), func(t *rapid.T) {
		// Generate random data
		existingDenoms := make(map[string]struct{})
		gcGen := rapid.Custom[sdk.Coin](func(t *rapid.T) sdk.Coin {
			return sdk.Coin{
				Denom: rapid.StringOfN(rapid.RuneFrom([]rune{'a', 'b', 'c'}), 5, 100, -1).
					Filter(func(s string) bool {
						_, ok := existingDenoms[s]
						existingDenoms[s] = struct{}{}
						return !ok
					}).
					Draw(t, "denom"),
				Amount: math.NewInt(rapid.Int64Range(1, 100_000).Draw(t, "coins")),
			}
		})
		gaugeCoins := sdk.NewCoins(rapid.SliceOfN[sdk.Coin](gcGen, 1, 100_000).Draw(t, "gaugeCoins")...)
		coinsToAdd := sdk.NewCoins(rapid.SliceOfN[sdk.Coin](gcGen, 1, 100_000).Draw(t, "coinsToAdd")...)

		s.SetupTest()

		// Create a new gauge
		_, _, existingGaugeCoins, _ := s.SetupNewGauge(true, gaugeCoins)
		owner := apptesting.CreateRandomAccounts(1)[0]
		// Fund the owner account
		s.FundAcc(owner, coinsToAdd)

		// Save the gas meter before the method call
		existingGasConsumed := s.Ctx.GasMeter().GasConsumed()

		// AddToGaugeRewards
		err := s.App.IncentivesKeeper.AddToGaugeRewards(s.Ctx, owner, coinsToAdd, 1)
		s.Require().NoError(err)

		// Min expected gas consumed
		baseGasFee := s.App.IncentivesKeeper.GetParams(s.Ctx).BaseGasFeeForAddRewardToGauge
		minimumGasConsumed := baseGasFee * uint64(len(gaugeCoins)+len(coinsToAdd))

		// Ensure that at least the minimum amount of gas was charged (based on number of additional gauge coins)
		gasConsumed := s.Ctx.GasMeter().GasConsumed() - existingGasConsumed
		fmt.Println(gasConsumed, minimumGasConsumed)
		s.Require().True(gasConsumed >= minimumGasConsumed)

		// Existing coins gets added to the module when we create gauge and add to gauge
		expectedCoins := existingGaugeCoins.Add(coinsToAdd...)

		// Check module account balance, should go up
		balance := s.App.BankKeeper.GetAllBalances(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName))
		s.Require().Equal(expectedCoins, balance)

		// Check gauge coins should go up
		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, 1)
		s.Require().NoError(err)

		s.Require().Equal(expectedCoins, gauge.Coins)
	})
}
