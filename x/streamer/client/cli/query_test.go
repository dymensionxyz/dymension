package cli_test

import (
	gocontext "context"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var defaultDistrInfo []types.DistrRecord = []types.DistrRecord{
	{
		GaugeId: 1,
		Weight:  math.NewInt(50),
	},
	{
		GaugeId: 2,
		Weight:  math.NewInt(50),
	},
}

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

// CreateStream creates a stream struct given the required params.
func (suite *QueryTestSuite) CreateStream(distrTo []types.DistrRecord, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

func (suite *QueryTestSuite) CreateDefaultStream(coins sdk.Coins) (uint64, *types.Stream) {
	return suite.CreateStream(defaultDistrInfo, coins, time.Now(), "day", 30)
}

func (suite *QueryTestSuite) SetupSuite() {
	suite.Setup()
	streamerCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500)), sdk.NewCoin("udym", sdk.NewInt(2500)))
	suite.FundModuleAcc(types.ModuleName, streamerCoins)
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	err := suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	testCases := []struct {
		name   string
		query  string
		input  interface{}
		output interface{}
	}{
		{
			"Query active streams",
			"/dymensionxyz.dymension.streamer.Query/ActiveStreams",
			&types.ActiveStreamsRequest{},
			&types.ActiveStreamsResponse{},
		},
		{
			"Query stream by id",
			"/dymensionxyz.dymension.streamer.Query/StreamByID",
			&types.StreamByIDRequest{Id: 1},
			&types.StreamByIDResponse{},
		},
		{
			"Query all streams",
			"/dymensionxyz.dymension.streamer.Query/Streams",
			&types.StreamsRequest{},
			&types.StreamsResponse{},
		},
		{
			"Query module to distibute coins",
			"/dymensionxyz.dymension.streamer.Query/ModuleToDistributeCoins",
			&types.ModuleToDistributeCoinsRequest{},
			&types.ModuleToDistributeCoinsResponse{},
		},
		{
			"Query upcoming streams",
			"/dymensionxyz.dymension.streamer.Query/UpcomingStreams",
			&types.UpcomingStreamsRequest{},
			&types.UpcomingStreamsResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupSuite()
			s.CreateDefaultStream(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500))))
			err := s.QueryHelper.Invoke(gocontext.Background(), tc.query, tc.input, tc.output)
			s.Require().NoError(err)
			s.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (suite *QueryTestSuite) CreateGauge() error {
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
