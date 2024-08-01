package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

const (
	Sponsored    = true
	NonSponsored = false
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

// SetupTest sets streamer parameters from the suite's context.
func (s *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(s.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.SponsorshipKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.App = app
	s.Ctx = ctx
	s.queryClient = queryClient
}

func (s *KeeperTestSuite) CreateGauge() {
	_, err := s.App.IncentivesKeeper.CreateGauge(
		s.Ctx,
		true,
		s.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByTime,
			Denom:         "stake",
			Duration:      time.Hour,
			Timestamp:     time.Time{},
		},
		time.Now(),
		1,
	)
	s.Require().NoError(err)
}

// CreateStream creates a stream given the required params.
func (s *KeeperTestSuite) CreateStream(
	distrTo []streamertypes.DistrRecord,
	coins sdk.Coins,
	startTime time.Time,
	epochIdetifier string,
	numEpoch uint64,
	sponsored bool,
) *streamertypes.Stream {
	s.FundModuleAcc(streamertypes.ModuleName, coins)
	streamID, err := s.App.StreamerKeeper.CreateStream(s.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch, sponsored)
	s.Require().NoError(err)
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)
	return stream
}

func (s *KeeperTestSuite) CreateDefaultSponsoredStream(coins sdk.Coins) *streamertypes.Stream {
	var defaultDistrInfo = []streamertypes.DistrRecord{
		{GaugeId: 1, Weight: math.NewInt(30)},
		{GaugeId: 2, Weight: math.NewInt(30)},
		{GaugeId: 3, Weight: math.NewInt(40)},
	}
	return s.CreateStream(defaultDistrInfo, coins, time.Now().Add(-1*time.Minute), "day", 30, Sponsored)
}

func (s *KeeperTestSuite) GetDistribution() types.Distribution {
	resp, err := s.queryClient.Distribution(s.Ctx, new(types.QueryDistributionRequest))
	s.Require().NoError(err)
	return resp.Distribution
}

func (s *KeeperTestSuite) GetVote(voter string) types.Vote {
	resp, err := s.queryClient.Vote(s.Ctx, &types.QueryVoteRequest{Voter: voter})
	s.Require().NoError(err)
	return resp.Vote
}

func (s *KeeperTestSuite) GetParams() types.Params {
	resp, err := s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)
	return resp.Params
}
