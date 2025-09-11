package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
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

	// fund streamer module
	streamerCoins := sdk.NewCoins(
		common.DymUint64(1000),
	)
	suite.FundModuleAcc(streamertypes.ModuleName, streamerCoins)
	suite.querier = keeper.NewQueryServerImpl(*suite.App.OTCBuybackKeeper)
}

func (suite *KeeperTestSuite) CreateDefaultAuction() uint64 {
	suite.FundModuleAcc(types.ModuleName, sdk.NewCoins(common.DymUint64(100)))

	// Create auction
	auctionID, err := suite.App.OTCBuybackKeeper.CreateAuction(suite.Ctx,
		common.DymUint64(100),
		suite.Ctx.BlockTime(),
		suite.Ctx.BlockTime().Add(24*time.Hour),
		math.LegacyNewDecWithPrec(5, 2),  // 5%
		math.LegacyNewDecWithPrec(50, 2), // 50%
		types.DefaultVestingParams,
		types.DefaultPumpParams,
	)
	suite.Require().NoError(err)
	return auctionID
}
