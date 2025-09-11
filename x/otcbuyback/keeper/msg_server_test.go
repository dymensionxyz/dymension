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

	var validCreateAuctionMsg = &types.MsgCreateAuction{
		Authority:       authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Allocation:      common.DymUint64(33), // streamer funded with 100 DYM on setup
		StartTime:       time.Now().Add(time.Hour),
		EndTime:         time.Now().Add(25 * time.Hour),   // 24 hour auction
		InitialDiscount: math.LegacyNewDecWithPrec(5, 2),  // 5%
		MaxDiscount:     math.LegacyNewDecWithPrec(50, 2), // 50%
		VestingParams: types.Auction_VestingParams{
			VestingPeriod:               24 * time.Hour,
			VestingStartAfterAuctionEnd: time.Hour,
		},
		PumpParams: types.Auction_PumpParams{
			StartTimeAfterAuctionEnd: time.Hour,
			EpochIdentifier:          "day",
			NumEpochsPaidOver:        30,
			NumOfPumpsPerEpoch:       1,
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
			name:        "success - valid auction creation",
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
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // Reset state for each test

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
