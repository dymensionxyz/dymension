package cli_test

import (
	"context"
	"strings"
	"testing"
	"time"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollapp "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryHelper *baseapp.QueryServiceTestHelper
}

// SetupLockAndGauge creates both a lock and a gauge.
func (suite *QueryTestSuite) CreateDefaultRollapp() string {
	alice := sdk.AccAddress("addr1---------------")
	// suite.FundAcc(alice, sdk.NewCoins(rollapptypes.DefaultRegistrationFee)) TODO: enable after x/dymns hooks are wired

	msgCreateRollapp := rollapptypes.MsgCreateRollapp{
		Creator:      alice.String(),
		RollappId:    urand.RollappID(),
		Bech32Prefix: strings.ToLower(tmrand.Str(3)),
		Alias:        strings.ToLower(tmrand.Str(3)),
		VmType:       rollapptypes.Rollapp_EVM,
	}

	msgServer := rollapp.NewMsgServerImpl(*suite.App.RollappKeeper)
	_, err := msgServer.CreateRollapp(suite.Ctx, &msgCreateRollapp)
	suite.Require().NoError(err)
	return msgCreateRollapp.RollappId
}

func (suite *QueryTestSuite) SetupSuite() {
	suite.App = apptesting.Setup(suite.T(), false)
	suite.Ctx = suite.App.BaseApp.NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQuerier(*suite.App.IncentivesKeeper))
	suite.queryHelper = queryHelper

	suite.CreateDefaultRollapp()
}

func (suite *QueryTestSuite) TestQueriesNeverAlterState() {
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
			"Query rollapp gauges",
			"/dymensionxyz.dymension.incentives.Query/RollappGauges",
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
			"Query module to distribute coins",
			"/dymensionxyz.dymension.incentives.Query/ModuleToDistributeCoins",
			&types.ModuleToDistributeCoinsRequest{},
			&types.ModuleToDistributeCoinsResponse{},
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

		suite.Run(tc.name, func() {
			suite.SetupSuite()
			err := suite.queryHelper.Invoke(context.Background(), tc.query, tc.input, tc.output)
			suite.Require().NoError(err)
			suite.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
