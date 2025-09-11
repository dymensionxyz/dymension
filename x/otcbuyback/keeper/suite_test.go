package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/keeper"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

const (
	Sponsored    = true
	NonSponsored = false
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	querier types.QueryServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.App = apptesting.Setup(suite.T())
	suite.Ctx = suite.App.NewContext(false).WithBlockTime(time.Now())
	suite.querier = keeper.NewQueryServerImpl(*suite.App.OTCBuybackKeeper)

	// Note: USDC has 6 decimals instead of 18
	// price is 4:1
	poolID := suite.PreparePoolWithCoins(sdk.NewCoins(
		sdk.NewCoin("usdc", math.NewInt(500_000).MulRaw(1e6)),
		sdk.NewCoin("adym", math.NewInt(2_000_000).MulRaw(1e18)),
	))

	spotPrice, err := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolID, "usdc", "adym")
	suite.Require().NoError(err)

	err = suite.App.OTCBuybackKeeper.SetAcceptedToken(suite.Ctx, "usdc", types.TokenData{
		PoolId:           poolID,
		LastAveragePrice: spotPrice,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) CreateDefaultAuction() uint64 {
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(100)))

	// Create default vesting and pump params
	vestingParams := types.Auction_VestingParams{
		VestingPeriod:               24 * time.Hour,
		VestingStartAfterAuctionEnd: 0,
	}
	pumpParams := types.Auction_PumpParams{
		StartTimeAfterAuctionEnd: time.Hour,
		EpochIdentifier:          "day",
		NumEpochsPaidOver:        30,
		NumOfPumpsPerEpoch:       1,
	}

	// Create auction
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
		common.DymUint64(100),
		suite.Ctx.BlockTime(),
		suite.Ctx.BlockTime().Add(24*time.Hour),
		math.LegacyNewDecWithPrec(10, 2), // 10%
		math.LegacyNewDecWithPrec(50, 2), // 50%
		vestingParams,
		pumpParams,
	)
	suite.Require().NoError(err)
	return auctionID
}

func (suite *KeeperTestSuite) CreateAuctionWithParams(vestingParams types.Auction_VestingParams, pumpParams types.Auction_PumpParams) uint64 {
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(100)))

	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
		common.DymUint64(100),
		suite.Ctx.BlockTime(),
		suite.Ctx.BlockTime().Add(24*time.Hour),
		math.LegacyNewDecWithPrec(5, 2),  // 5%
		math.LegacyNewDecWithPrec(50, 2), // 50%
		vestingParams,
		pumpParams,
	)
	suite.Require().NoError(err)
	return auctionID
}

// CreateRandomAccount creates a random account address for testing
func (suite *KeeperTestSuite) CreateRandomAccount() sdk.AccAddress {
	return sample.Acc()
}

// BuySomeTokens is a helper function that executes a buy transaction and returns the payment coin
func (suite *KeeperTestSuite) BuySomeTokens(auctionID uint64, buyer sdk.AccAddress, amountToBuy math.Int, denomToPay string) sdk.Coin {
	paymentCoin, err := suite.App.OTCBuybackKeeper.Buy(suite.Ctx, buyer, auctionID, amountToBuy, denomToPay)
	suite.Require().NoError(err, "Buy operation should succeed")
	suite.Require().True(paymentCoin.IsPositive(), "Payment coin should be positive")
	return paymentCoin
}
