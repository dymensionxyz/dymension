package keeper_test

import (
	"time"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	keeper "github.com/dymensionxyz/dymension/x/streamer/keeper"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

var (
	defaultDestAddr sdk.AccAddress = sdk.AccAddress([]byte("addr1---------------"))
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	moduleAddress sdk.AccAddress

	querier keeper.Querier
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.moduleAddress = authtypes.NewModuleAddress(types.ModuleName)
	streamerCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500)), sdk.NewCoin("udym", sdk.NewInt(2500)))
	suite.FundModuleAcc(types.ModuleName, streamerCoins)
	suite.querier = keeper.NewQuerier(suite.App.StreamerKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// CreateStream creates a stream struct given the required params.
func (suite *KeeperTestSuite) CreateStream(distrTo sdk.AccAddress, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

func (suite *KeeperTestSuite) CreateDefaultStream(coins sdk.Coins) (uint64, *types.Stream) {
	return suite.CreateStream(defaultDestAddr, coins, time.Time{}, "day", 30)
}

func (suite *KeeperTestSuite) ExpectedDefaultStream(streamID uint64, coins sdk.Coins) types.Stream {
	return types.Stream{
		Id:                   streamID,
		DistributeTo:         defaultDestAddr.String(),
		Coins:                coins,
		StartTime:            time.Time{},
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    30,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
	}
}
