package cli_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// create a stream
	_, err := s.App.StreamerKeeper.CreateStream(
		s.Ctx,
		true,
		s.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         gammtypes.GetPoolShareDenom(poolID),
			Duration:      time.Hour,
			Timestamp:     time.Time{},
		},
		s.Ctx.BlockTime(),
		1,
	)
	s.NoError(err)
	s.Commit()
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
			"/osmosis.incentives.Query/ActiveStreams",
			&types.ActiveStreamsRequest{},
			&types.ActiveStreamsResponse{},
		},
		{
			"Query active streams per denom",
			"/osmosis.incentives.Query/ActiveStreamsPerDenom",
			&types.ActiveStreamsPerDenomRequest{Denom: "stake"},
			&types.ActiveStreamsPerDenomResponse{},
		},
		{
			"Query stream by id",
			"/osmosis.incentives.Query/StreamByID",
			&types.StreamByIDRequest{Id: 1},
			&types.StreamByIDResponse{},
		},
		{
			"Query all streams",
			"/osmosis.incentives.Query/Streams",
			&types.StreamsRequest{},
			&types.StreamsResponse{},
		},
		{
			"Query lockable durations",
			"/osmosis.incentives.Query/LockableDurations",
			&types.QueryLockableDurationsRequest{},
			&types.QueryLockableDurationsResponse{},
		},
		{
			"Query module to distibute coins",
			"/osmosis.incentives.Query/ModuleToDistributeCoins",
			&types.ModuleToDistributeCoinsRequest{},
			&types.ModuleToDistributeCoinsResponse{},
		},
		{
			"Query reward estimate",
			"/osmosis.incentives.Query/RewardsEst",
			&types.RewardsEstRequest{Owner: s.TestAccs[0].String()},
			&types.RewardsEstResponse{},
		},
		{
			"Query upcoming streams",
			"/osmosis.incentives.Query/UpcomingStreams",
			&types.UpcomingStreamsRequest{},
			&types.UpcomingStreamsResponse{},
		},
		{
			"Query upcoming streams",
			"/osmosis.incentives.Query/UpcomingStreamsPerDenom",
			&types.UpcomingStreamsPerDenomRequest{Denom: "stake"},
			&types.UpcomingStreamsPerDenomResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupSuite()
			err := s.QueryHelper.Invoke(gocontext.Background(), tc.query, tc.input, tc.output)
			s.Require().NoError(err)
			s.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
