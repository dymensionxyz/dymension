package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	msgServer   keeper.MsgServer
}

// SetupTest sets streamer parameters from the suite's context.
func (s *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(s.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.SponsorshipKeeper))
	queryClient := types.NewQueryClient(queryHelper)
	msgServer := keeper.NewMsgServer(app.SponsorshipKeeper)

	s.App = app
	s.Ctx = ctx
	s.queryClient = queryClient
	s.msgServer = msgServer
}

func (s *KeeperTestSuite) CreateGauge() uint64 {
	s.T().Helper()

	gaugeID, err := s.App.IncentivesKeeper.CreateGauge(
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
	return gaugeID
}

func (s *KeeperTestSuite) CreateGauges(num int) {
	s.T().Helper()

	for i := 0; i < num; i++ {
		s.CreateGauge()
	}
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
	s.T().Helper()

	s.FundModuleAcc(streamertypes.ModuleName, coins)
	streamID, err := s.App.StreamerKeeper.CreateStream(s.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch, sponsored)
	s.Require().NoError(err)
	stream, err := s.App.StreamerKeeper.GetStreamByID(s.Ctx, streamID)
	s.Require().NoError(err)
	return stream
}

func (s *KeeperTestSuite) CreateDefaultSponsoredStream(coins sdk.Coins) *streamertypes.Stream {
	s.T().Helper()

	defaultDistrInfo := []streamertypes.DistrRecord{
		{GaugeId: 1, Weight: math.NewInt(30)},
		{GaugeId: 2, Weight: math.NewInt(30)},
		{GaugeId: 3, Weight: math.NewInt(40)},
	}
	return s.CreateStream(defaultDistrInfo, coins, time.Now().Add(-1*time.Minute), "day", 30, Sponsored)
}

func (s *KeeperTestSuite) GetDistribution() types.Distribution {
	s.T().Helper()

	resp, err := s.queryClient.Distribution(s.Ctx, new(types.QueryDistributionRequest))
	s.Require().NoError(err)
	return resp.Distribution
}

func (s *KeeperTestSuite) GetVote(voter string) types.Vote {
	s.T().Helper()

	resp, err := s.queryClient.Vote(s.Ctx, &types.QueryVoteRequest{Voter: voter})
	s.Require().NoError(err)
	return resp.Vote
}

func (s *KeeperTestSuite) GetParams() types.Params {
	s.T().Helper()

	resp, err := s.queryClient.Params(s.Ctx, new(types.QueryParamsRequest))
	s.Require().NoError(err)
	return resp.Params
}

func (s *KeeperTestSuite) Vote(vote types.MsgVote) {
	s.T().Helper()

	voteResp, err := s.msgServer.Vote(s.Ctx, &vote)
	s.Require().NoError(err)
	s.Require().NotNil(voteResp)
}

func (s *KeeperTestSuite) CreateValidator() stakingtypes.ValidatorI {
	s.T().Helper()

	valAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, sdk.NewInt(1_000_000_000))

	// Build MsgCreateValidator
	valAddr := sdk.ValAddress(valAddrs[0].Bytes())
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1_000_000_000)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
	)
	s.Require().NoError(err)

	// Create a validator
	handler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgCreateValidator))
	resp, err := handler(s.Ctx, msgCreate)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	val, found := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().True(found)

	return val
}

func (s *KeeperTestSuite) CreateDelegator(valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	delAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, sdk.NewInt(1_000_000_000))
	delAddr := delAddrs[0]
	return s.Delegate(delAddr, valAddr, coin)
}

func (s *KeeperTestSuite) Delegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	handler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgDelegate))
	resp, err := handler(s.Ctx, stakingtypes.NewMsgDelegate(delAddr, valAddr, coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	del, found := s.App.StakingKeeper.GetDelegation(s.Ctx, delAddr, valAddr)
	s.Require().True(found)

	return del
}

// Undelegate sends MsgUndelegate and returns the delegation object. Return value might me nil in case if
// the delegator completely unbonds.
func (s *KeeperTestSuite) Undelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	handler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgUndelegate))
	resp, err := handler(s.Ctx, stakingtypes.NewMsgUndelegate(delAddr, valAddr, coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	return s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valAddr)
}

// Undelegate sends MsgUndelegate and returns the delegation object. Src return value might me nil in case if
// the delegator completely unbonds.
func (s *KeeperTestSuite) BeginRedelegate(
	delAddr sdk.AccAddress,
	valSrcAddr, valDstAddr sdk.ValAddress,
	coin sdk.Coin,
) (src, dst stakingtypes.DelegationI) {
	s.T().Helper()

	handler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgBeginRedelegate))
	resp, err := handler(s.Ctx, stakingtypes.NewMsgBeginRedelegate(delAddr, valSrcAddr, valDstAddr, coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	return s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valSrcAddr),
		s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valDstAddr)
}

func (s *KeeperTestSuite) CancelUnbondingDelegation(
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	creationHeight int64,
	coin sdk.Coin,
) stakingtypes.DelegationI {
	s.T().Helper()

	handler := s.App.MsgServiceRouter().Handler(new(stakingtypes.MsgCancelUnbondingDelegation))
	resp, err := handler(s.Ctx, stakingtypes.NewMsgCancelUnbondingDelegation(delAddr, valAddr, creationHeight, coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	src, found := s.App.StakingKeeper.GetDelegation(s.Ctx, delAddr, valAddr)
	s.Require().True(found)

	return src
}

func accAddrsToString(a []sdk.AccAddress) []string {
	res := make([]string, 0, len(a))
	for _, addr := range a {
		res = append(res, addr.String())
	}
	return res
}
