package cli_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/incentives/types"
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

	// create a pool
	poolID := s.PrepareBalancerPool()

	// set up lock with id = 1
	s.LockTokens(s.TestAccs[0], sdk.Coins{sdk.NewCoin("gamm/pool/1", sdk.NewInt(1000000))}, time.Hour*24)

	// create a gauge
	_, err := s.App.IncentivesKeeper.CreateGauge(
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
			"Query active gauges",
			"/dymensionxyz.dymension.incentives.Query/ActiveGauges",
			&types.ActiveGaugesRequest{},
			&types.ActiveGaugesResponse{},
		},
		{
			"Query active gauges per denom",
			"/dymensionxyz.dymension.incentives.Query/ActiveGaugesPerDenom",
			&types.ActiveGaugesPerDenomRequest{Denom: "stake"},
			&types.ActiveGaugesPerDenomResponse{},
		},
		{
			"Query gauge by id",
			"/dymensionxyz.dymension.incentives.Query/GaugeByID",
			&types.GaugeByIDRequest{Id: 1},
			&types.GaugeByIDResponse{},
		},
		{
			"Query all gauges",
			"/dymensionxyz.dymension.incentives.Query/Gauges",
			&types.GaugesRequest{},
			&types.GaugesResponse{},
		},
		{
			"Query lockable durations",
			"/dymensionxyz.dymension.incentives.Query/LockableDurations",
			&types.QueryLockableDurationsRequest{},
			&types.QueryLockableDurationsResponse{},
		},
		{
			"Query module to distibute coins",
			"/dymensionxyz.dymension.incentives.Query/ModuleToDistributeCoins",
			&types.ModuleToDistributeCoinsRequest{},
			&types.ModuleToDistributeCoinsResponse{},
		},
		{
			"Query reward estimate",
			"/dymensionxyz.dymension.incentives.Query/RewardsEst",
			&types.RewardsEstRequest{Owner: s.TestAccs[0].String()},
			&types.RewardsEstResponse{},
		},
		{
			"Query upcoming gauges",
			"/dymensionxyz.dymension.incentives.Query/UpcomingGauges",
			&types.UpcomingGaugesRequest{},
			&types.UpcomingGaugesResponse{},
		},
		{
			"Query upcoming gauges",
			"/dymensionxyz.dymension.incentives.Query/UpcomingGaugesPerDenom",
			&types.UpcomingGaugesPerDenomRequest{Denom: "stake"},
			&types.UpcomingGaugesPerDenomResponse{},
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
