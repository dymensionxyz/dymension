package keeper_test

import (
	"slices"
	"testing"
	"time"

	"cosmossdk.io/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	sponsorshipkeeper "github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

const (
	Sponsored    = true
	NonSponsored = false
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

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	querier keeper.Querier
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.App = apptesting.Setup(suite.T(), false)
	suite.Ctx = suite.App.BaseApp.NewContext(false, cometbftproto.Header{Height: 1, ChainID: "dymension_100-1", Time: time.Now().UTC()})
	streamerCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2500000)), sdk.NewCoin("udym", sdk.NewInt(2500000)))
	suite.FundModuleAcc(types.ModuleName, streamerCoins)
	suite.querier = keeper.NewQuerier(suite.App.StreamerKeeper)

	err := suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)
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

// CreateStream creates a non-sponsored stream struct given the required params.
func (suite *KeeperTestSuite) CreateStream(distrTo []types.DistrRecord, coins sdk.Coins, startTime time.Time, epochIdentifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdentifier, numEpoch, NonSponsored)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

// CreateStream creates a sponsored stream struct given the required params.
func (suite *KeeperTestSuite) CreateSponsoredStream(distrTo []types.DistrRecord, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch, Sponsored)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

func (suite *KeeperTestSuite) CreateDefaultStream(coins sdk.Coins) (uint64, *types.Stream) {
	return suite.CreateStream(defaultDistrInfo, coins, time.Now().Add(-1*time.Minute), "day", 30)
}

func (suite *KeeperTestSuite) ExpectedDefaultStream(streamID uint64, starttime time.Time, coins sdk.Coins) types.Stream {
	distInfo, err := types.NewDistrInfo(defaultDistrInfo)
	suite.Require().NoError(err)

	const numEpochsPaidOver = 30
	return types.Stream{
		Id:                   streamID,
		DistributeTo:         distInfo,
		Coins:                coins,
		StartTime:            starttime,
		DistrEpochIdentifier: "day",
		NumEpochsPaidOver:    numEpochsPaidOver,
		FilledEpochs:         0,
		DistributedCoins:     sdk.Coins{},
		Sponsored:            false,
		EpochCoins:           coins.QuoInt(math.NewInt(numEpochsPaidOver)),
	}
}

func (suite *KeeperTestSuite) CreateGauges(num int) {
	suite.T().Helper()

	for i := 0; i < num; i++ {
		err := suite.CreateGauge()
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) CreateGaugesUntil(num int) {
	suite.T().Helper()

	gauges := suite.App.IncentivesKeeper.GetGauges(suite.Ctx)
	remain := num - len(gauges)

	for i := 0; i < remain; i++ {
		err := suite.CreateGauge()
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) Distribution() sponsorshiptypes.Distribution {
	queryServer := sponsorshipkeeper.NewQueryServer(suite.App.SponsorshipKeeper)
	d, err := queryServer.Distribution(suite.Ctx, new(sponsorshiptypes.QueryDistributionRequest))
	suite.Require().NoError(err)
	suite.Require().NotNil(d)
	return d.Distribution
}

// Vote creates two validators and a delegator, then delegates the stake to these validators.
// The delegator then casts the vote to gauges through x/sponsorship.
func (suite *KeeperTestSuite) Vote(vote sponsorshiptypes.MsgVote, votingPower math.Int) {
	suite.T().Helper()

	val1 := suite.CreateValidator()
	val2 := suite.CreateValidator()

	delAddr, err := sdk.AccAddressFromBech32(vote.Voter)
	suite.Require().NoError(err)
	initialBalance := sdk.NewCoin(sdk.DefaultBondDenom, votingPower)
	apptesting.FundAccount(suite.App, suite.Ctx, delAddr, sdk.NewCoins(initialBalance))

	stake := votingPower.Quo(math.NewInt(2))
	delegation := sdk.NewCoin(sdk.DefaultBondDenom, stake)
	suite.Delegate(delAddr, val1.GetOperator(), delegation) // delegator 1 -> validator 1
	suite.Delegate(delAddr, val2.GetOperator(), delegation) // delegator 1 -> validator 2

	suite.vote(vote)
}

func (suite *KeeperTestSuite) vote(vote sponsorshiptypes.MsgVote) {
	suite.T().Helper()

	msgServer := sponsorshipkeeper.NewMsgServer(suite.App.SponsorshipKeeper)
	voteResp, err := msgServer.Vote(suite.Ctx, &vote)
	suite.Require().NoError(err)
	suite.Require().NotNil(voteResp)
}

func (suite *KeeperTestSuite) CreateValidator() stakingtypes.ValidatorI {
	suite.T().Helper()

	valAddrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, sdk.NewInt(1_000_000_000))

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
	suite.Require().NoError(err)

	// Create a validator
	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(suite.App.StakingKeeper)
	resp, err := stakingMsgSrv.CreateValidator(suite.Ctx, msgCreate)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	val, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddr)
	suite.Require().True(found)

	return val
}

func (suite *KeeperTestSuite) CreateDelegator(valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	suite.T().Helper()

	delAddrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, sdk.NewInt(1_000_000_000))
	delAddr := delAddrs[0]
	return suite.Delegate(delAddr, valAddr, coin)
}

func (suite *KeeperTestSuite) Delegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	suite.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(suite.App.StakingKeeper)
	resp, err := stakingMsgSrv.Delegate(suite.Ctx, stakingtypes.NewMsgDelegate(delAddr, valAddr, coin))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	del, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, delAddr, valAddr)
	suite.Require().True(found)

	return del
}

func (suite *KeeperTestSuite) DistributeAllRewards(streams []types.Stream) sdk.Coins {
	rewards := sdk.Coins{}
	suite.Require().True(slices.IsSortedFunc(streams, keeper.CmpStreams))
	for _, stream := range streams {
		res := suite.App.StreamerKeeper.DistributeRewards(
			suite.Ctx,
			types.NewEpochPointer(stream.DistrEpochIdentifier),
			types.IterationsNoLimit,
			[]types.Stream{stream},
		)
		suite.Require().Len(res.FilledStreams, 1)
		err := suite.App.StreamerKeeper.SetStream(suite.Ctx, &res.FilledStreams[0])
		suite.Require().NoError(err)
		rewards = rewards.Add(res.DistributedCoins...)
	}
	return rewards
}
