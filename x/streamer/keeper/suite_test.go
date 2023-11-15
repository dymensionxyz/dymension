package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"testing"

	"github.com/stretchr/testify/suite"

	keeper "github.com/dymensionxyz/dymension/x/streamer/keeper"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

var (
	defaultDistrInfo *types.DistrInfo = &types.DistrInfo{
		Name:        "",
		TotalWeight: math.NewInt(100),
		Records: []types.DistrRecord{{
			GaugeId: 1,
			Weight:  math.NewInt(50),
		},
			{
				GaugeId: 2,
				Weight:  math.NewInt(50),
			},
		},
	}
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	querier keeper.Querier
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	streamerCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500000)), sdk.NewCoin("udym", sdk.NewInt(2500000)))
	suite.FundModuleAcc(types.ModuleName, streamerCoins)
	suite.querier = keeper.NewQuerier(suite.App.StreamerKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) CreateGauge() error {
	_, err := suite.App.IncentivesKeeper.CreateGauge(
		suite.Ctx,
		true,
		suite.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByTime,
			Denom:         "stake",
			Duration:      time.Hour,
			Timestamp:     time.Time{},
		}, time.Now(), 1)
	return err
}

// CreateStream creates a stream struct given the required params.
func (suite *KeeperTestSuite) CreateStream(distrTo *types.DistrInfo, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

func (suite *KeeperTestSuite) CreateDefaultStream(coins sdk.Coins) (uint64, *types.Stream) {
	return suite.CreateStream(defaultDistrInfo, coins, time.Now().Add(-1*time.Minute), "day", 30)
}

func (suite *KeeperTestSuite) ExpectedDefaultStream(streamID uint64, starttime time.Time, coins sdk.Coins) types.Stream {
	return types.Stream{
		Id:                   streamID,
		DistributeTo:         defaultDistrInfo,
		Coins:                coins,
		StartTime:            starttime,
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    30,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
	}
}
