package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/keeper"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func (suite *KeeperTestSuite) TestMsgServer_CreateAuction() {
	var tcMsg types.MsgCreateAuction

	linearDiscount := types.NewLinearDiscountType(
		math.LegacyNewDecWithPrec(2, 1), // 0.2 = 20% initial discount
		math.LegacyNewDecWithPrec(5, 1), // 0.5 = 50% max discount
		24*time.Hour,
	)

	validCreateAuctionMsg := &types.MsgCreateAuction{
		Authority:    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Allocation:   common.DymUint64(100),
		StartTime:    time.Now().Add(time.Hour),
		EndTime:      time.Now().Add(25 * time.Hour), // 24 hour auction
		DiscountType: linearDiscount,
		VestingDelay: time.Hour,
		PumpParams: types.Auction_PumpParams{
			EpochIdentifier:    "day",
			NumEpochs:          30,
			NumOfPumpsPerEpoch: 1,
			PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
			PumpDelay:          time.Hour,
			PumpInterval:       time.Hour,
		},
	}

	testCases := []struct {
		name        string
		setup       func()
		expectError bool
		postCheck   func()
	}{
		{
			name: "error - invalid authority",
			setup: func() {
				tcMsg.Authority = sdk.AccAddress([]byte("invalid_authority")).String()
			},
			expectError: true,
		},
		{
			name: "error - insufficient streamer funds",
			setup: func() {
				tcMsg.Allocation = common.DymUint64(1000000000000)
			},
			expectError: true,
		},
		{
			name: "error - invalid allocation denom",
			setup: func() {
				invalid_coin := sdk.NewCoin("invalid_denom", math.NewIntWithDecimal(1, 18))
				streamerCoins := sdk.NewCoins(invalid_coin)
				suite.FundModuleAcc(streamertypes.ModuleName, streamerCoins)

				tcMsg.Allocation = invalid_coin
			},
			expectError: true,
		},
		{
			name:        "success - valid linear discount auction",
			setup:       func() {},
			expectError: false,
			postCheck: func() {
				// Verify auction was created
				auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, 1)
				suite.Require().True(found, "auction should be created")
				suite.Require().False(auction.IsCompleted(), "new auction should not be completed")

				// Verify funds were transferred from streamer to otcbuyback
				otcBalance := suite.App.BankKeeper.GetBalance(suite.Ctx,
					authtypes.NewModuleAddress(types.ModuleName), common.DYMCoin.Denom)
				suite.Require().Equal(validCreateAuctionMsg.Allocation.String(), otcBalance.String(),
					"otcbuyback should have received allocation funds")
			},
		},
		{
			name: "success - valid fixed discount auction",
			setup: func() {
				tcMsg.DiscountType = types.NewFixedDiscountType([]types.FixedDiscount_Discount{
					{Discount: math.LegacyNewDecWithPrec(10, 2), VestingPeriod: 30 * 24 * time.Hour},
					{Discount: math.LegacyNewDecWithPrec(30, 2), VestingPeriod: 90 * 24 * time.Hour},
					{Discount: math.LegacyNewDecWithPrec(50, 2), VestingPeriod: 180 * 24 * time.Hour},
				})
			},
			expectError: false,
			postCheck: func() {
				auction, found := suite.App.OTCBuybackKeeper.GetAuction(suite.Ctx, 1)
				suite.Require().True(found)
				suite.Require().NotNil(auction.DiscountType.GetFixed())
				suite.Require().Equal(3, len(auction.DiscountType.GetFixed().Discounts))
			},
		},
		{
			name: "error - fixed discount with empty discounts",
			setup: func() {
				tcMsg.DiscountType = types.NewFixedDiscountType([]types.FixedDiscount_Discount{})
			},
			expectError: true,
		},
		{
			name: "error - fixed discount with invalid discount rate",
			setup: func() {
				tcMsg.DiscountType = types.NewFixedDiscountType([]types.FixedDiscount_Discount{
					{Discount: math.LegacyNewDecWithPrec(15, 1), VestingPeriod: 30 * 24 * time.Hour}, // 1.5 = 150% invalid
				})
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset state for each test

			// fund streamer module
			suite.FundModuleAcc(streamertypes.ModuleName, sdk.NewCoins(common.DymUint64(100)))

			tcMsg = *validCreateAuctionMsg

			if tc.setup != nil {
				tc.setup()
			}

			msgServer := keeper.NewMsgServerImpl(*suite.App.OTCBuybackKeeper)
			resp, err := msgServer.CreateAuction(suite.Ctx, &tcMsg)

			if tc.expectError {
				suite.Require().Error(err, "expected error for test case: %s", tc.name)
			} else {
				suite.Require().NoError(err, "unexpected error for test case: %s", tc.name)
				suite.Require().Greater(resp.AuctionId, uint64(0), "auction ID should be positive")
				if tc.postCheck != nil {
					tc.postCheck()
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgServer_SetAcceptedTokens() {
	var tcMsg types.MsgSetAcceptedTokens

	// use func to allocate a new AcceptedTokens slice on every test rather than reusing the same one
	validMsg := func() types.MsgSetAcceptedTokens {
		return types.MsgSetAcceptedTokens{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			AcceptedTokens: []types.MsgSetAcceptedTokens_Token{
				{
					Denom:  "usdc",
					PoolId: 1,
				},
			},
		}
	}

	testCases := []struct {
		name        string
		setup       func()
		expectError bool
		postCheck   func()
	}{
		{
			name: "error - invalid authority",
			setup: func() {
				tcMsg.Authority = sdk.AccAddress([]byte("invalid_authority")).String()
			},
			expectError: true,
		},
		{
			name: "error - pool doesn't exist",
			setup: func() {
				tcMsg.AcceptedTokens[0].PoolId = 999
			},
			expectError: true,
		},
		{
			name: "error - pool doesn't have base denom",
			setup: func() {
				// Create a pool without the base denom (adym)
				poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
					sdk.NewCoin("usdc", math.NewInt(500_000).MulRaw(1e6)),
					sdk.NewCoin("atom", math.NewInt(1_000_000).MulRaw(1e6)),
				))
				tcMsg.AcceptedTokens[0].PoolId = poolID
			},
			expectError: true,
		},
		{
			name: "error - pool doesn't have token denom",
			setup: func() {
				// Create a pool without the token denom (usdc)
				poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
					sdk.NewCoin("atom", math.NewInt(1_000_000).MulRaw(1e6)),
					sdk.NewCoin("adym", math.NewInt(2_000_000).MulRaw(1e18)),
				))
				tcMsg.AcceptedTokens[0].PoolId = poolID
			},
			expectError: true,
		},
		{
			name: "success - valid pool with base denom first",
			setup: func() {
				// Create a pool with adym (base denom) first
				poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
					sdk.NewCoin("adym", math.NewInt(2_000_000).MulRaw(1e18)),
					sdk.NewCoin("usdc", math.NewInt(500_000).MulRaw(1e6)),
				))
				tcMsg.AcceptedTokens[0].PoolId = poolID
			},
			expectError: false,
			postCheck: func() {
				// Verify token was set
				tokenData, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
				suite.Require().NoError(err, "token should be set")
				suite.Require().Equal(tcMsg.AcceptedTokens[0].PoolId, tokenData.PoolId)
				suite.Require().True(tokenData.LastAveragePrice.IsPositive(), "price should be positive")
			},
		},
		{
			name: "success - valid pool with token denom first",
			setup: func() {
				// Create a pool with usdc (token denom) first
				poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
					sdk.NewCoin("usdc", math.NewInt(500_000).MulRaw(1e6)),
					sdk.NewCoin("adym", math.NewInt(2_000_000).MulRaw(1e18)),
				))
				tcMsg.AcceptedTokens[0].PoolId = poolID
			},
			expectError: false,
			postCheck: func() {
				// Verify token was set
				tokenData, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
				suite.Require().NoError(err, "token should be set")
				suite.Require().Equal(tcMsg.AcceptedTokens[0].PoolId, tokenData.PoolId)
				suite.Require().True(tokenData.LastAveragePrice.IsPositive(), "price should be positive")
			},
		},
		{
			name: "success - multiple tokens",
			setup: func() {
				// Create another pool with a different token
				poolID2 := suite.PreparePoolWithCoins(sdk.NewCoins(
					sdk.NewCoin("atom", math.NewInt(1_000_000).MulRaw(1e6)),
					sdk.NewCoin("adym", math.NewInt(4_000_000).MulRaw(1e18)),
				))
				tcMsg.AcceptedTokens = append(tcMsg.AcceptedTokens, types.MsgSetAcceptedTokens_Token{
					Denom:  "atom",
					PoolId: poolID2,
				})
			},
			expectError: false,
			postCheck: func() {
				// Verify both tokens were set
				tokenData1, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "usdc")
				suite.Require().NoError(err, "usdc token should be set")
				suite.Require().True(tokenData1.LastAveragePrice.IsPositive())

				tokenData2, err := suite.App.OTCBuybackKeeper.GetAcceptedTokenData(suite.Ctx, "atom")
				suite.Require().NoError(err, "atom token should be set")
				suite.Require().True(tokenData2.LastAveragePrice.IsPositive())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset state for each test

			tcMsg = validMsg()

			if tc.setup != nil {
				tc.setup()
			}

			msgServer := keeper.NewMsgServerImpl(*suite.App.OTCBuybackKeeper)
			_, err := msgServer.SetAcceptedTokens(suite.Ctx, &tcMsg)

			if tc.expectError {
				suite.Require().Error(err, "expected error for test case: %s", tc.name)
			} else {
				suite.Require().NoError(err, "unexpected error for test case: %s", tc.name)
				if tc.postCheck != nil {
					tc.postCheck()
				}
			}
		})
	}
}
