package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/keeper"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
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

func (suite *KeeperTestSuite) CreateDefaultLinearAuction() uint64 {
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(100)))

	// Create default vesting and pump params
	vestingParams := types.Auction_VestingParams{
		VestingDelay: 0,
	}
	pumpParams := types.Auction_PumpParams{
		EpochIdentifier:    "day",
		NumEpochs:          30,
		NumOfPumpsPerEpoch: 1,
		PumpDistr:          streamertypes.PumpDistr_PUMP_DISTR_UNIFORM,
		PumpDelay:          time.Hour,
		PumpInterval:       time.Hour,
	}

	discount := types.NewLinearDiscountType(
		math.LegacyNewDecWithPrec(1, 1), // 0.1 = 10% initial discount
		math.LegacyNewDecWithPrec(5, 1), // 0.5 = 50% max discount
		24*time.Hour,
	)

	// Create auction
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
		common.DymUint64(100),
		suite.Ctx.BlockTime(),
		suite.Ctx.BlockTime().Add(24*time.Hour),
		discount,
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
