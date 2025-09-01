package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	sponsorshipkeeper "github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

const (
	Sponsored    = true
	NonSponsored = false
)

var defaultDistrInfo = []types.DistrRecord{
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
	suite.App = apptesting.Setup(suite.T())
	suite.Ctx = suite.App.NewContext(false).WithBlockTime(time.Now())
	streamerCoins := sdk.NewCoins(
		sdk.NewCoin(sdk.DefaultBondDenom, common.DYM.MulRaw(100)),
		sdk.NewCoin("udym", math.NewInt(2500000)),
		common.DymUint64(100),
	)
	suite.FundModuleAcc(types.ModuleName, streamerCoins)
	suite.querier = keeper.NewQuerier(suite.App.StreamerKeeper)

	// Disable any distribution threshold for tests
	ip := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	bd, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	suite.Require().NoError(err)
	ip.MinValueForDistribution = sdk.NewCoin(bd, math.ZeroInt())
	suite.App.IncentivesKeeper.SetParams(suite.Ctx, ip)

	// Fund alice, the default rollapp creator, so she has enough balance for IRO creation
	funds := suite.App.IROKeeper.GetParams(suite.Ctx).CreationFee.Mul(math.NewInt(10)) // 10 times the creation fee
	suite.FundAcc(sdk.MustAccAddressFromBech32(apptesting.Alice), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds)))

	err = suite.CreateGauge()
	suite.Require().NoError(err)
	err = suite.CreateGauge()
	suite.Require().NoError(err)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) CreateGauge() error {
	_, err := suite.App.IncentivesKeeper.CreateAssetGauge(
		suite.Ctx,
		true,
		suite.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			Denom:    "stake",
			Duration: time.Hour,
		}, time.Now(), 1)
	return err
}

// CreateStream creates a non-sponsored stream struct given the required params.
func (suite *KeeperTestSuite) CreateStream(distrTo []types.DistrRecord, coins sdk.Coins, startTime time.Time, epochIdentifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdentifier, numEpoch, NonSponsored, nil)
	suite.Require().NoError(err)
	stream, err := suite.App.StreamerKeeper.GetStreamByID(suite.Ctx, streamID)
	suite.Require().NoError(err)
	return streamID, stream
}

// CreateSponsoredStream creates a sponsored stream struct given the required params.
func (suite *KeeperTestSuite) CreateSponsoredStream(distrTo []types.DistrRecord, coins sdk.Coins, startTime time.Time, epochIdetifier string, numEpoch uint64) (uint64, *types.Stream) {
	streamID, err := suite.App.StreamerKeeper.CreateStream(suite.Ctx, coins, distrTo, startTime, epochIdetifier, numEpoch, Sponsored, nil)
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
func (suite *KeeperTestSuite) CreateValVote(vote sponsorshiptypes.MsgVote, votingPower math.Int) {
	suite.T().Helper()

	val1 := suite.CreateValidator()
	val2 := suite.CreateValidator()

	val1Addr, err := sdk.ValAddressFromBech32(val1.GetOperator())
	suite.Require().NoError(err)

	val2Addr, err := sdk.ValAddressFromBech32(val2.GetOperator())
	suite.Require().NoError(err)

	delAddr, err := sdk.AccAddressFromBech32(vote.Voter)
	suite.Require().NoError(err)
	initialBalance := sdk.NewCoin(sdk.DefaultBondDenom, votingPower)
	apptesting.FundAccount(suite.App, suite.Ctx, delAddr, sdk.NewCoins(initialBalance))

	stake := votingPower.Quo(math.NewInt(2))
	delegation := sdk.NewCoin(sdk.DefaultBondDenom, stake)
	suite.Delegate(delAddr, val1Addr, delegation) // delegator 1 -> validator 1
	suite.Delegate(delAddr, val2Addr, delegation) // delegator 1 -> validator 2

	suite.Vote(vote)
}

func (suite *KeeperTestSuite) Vote(vote sponsorshiptypes.MsgVote) {
	suite.T().Helper()

	msgServer := sponsorshipkeeper.NewMsgServer(suite.App.SponsorshipKeeper)
	voteResp, err := msgServer.Vote(suite.Ctx, &vote)
	suite.Require().NoError(err)
	suite.Require().NotNil(voteResp)
}

func (suite *KeeperTestSuite) CreateValidator() stakingtypes.ValidatorI {
	suite.T().Helper()

	valAddrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1_000_000_000))

	// Build MsgCreateValidator
	valAddr := sdk.ValAddress(valAddrs[0].Bytes())
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000_000)),
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
		math.OneInt(),
	)
	suite.Require().NoError(err)

	// Create a validator
	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(suite.App.StakingKeeper)
	resp, err := stakingMsgSrv.CreateValidator(suite.Ctx, msgCreate)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	val, err := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddr)
	suite.Require().NoError(err)

	return val
}

func (suite *KeeperTestSuite) CreateDelegator(valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	suite.T().Helper()

	delAddrs := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, common.DYM.MulRaw(1_000))
	delAddr := delAddrs[0]
	return suite.Delegate(delAddr, valAddr, coin)
}

func (suite *KeeperTestSuite) Delegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	suite.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(suite.App.StakingKeeper)
	resp, err := stakingMsgSrv.Delegate(suite.Ctx, stakingtypes.NewMsgDelegate(delAddr.String(), valAddr.String(), coin))
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	del, err := suite.App.StakingKeeper.GetDelegation(suite.Ctx, delAddr, valAddr)
	suite.Require().NoError(err)

	return del
}

func (suite *KeeperTestSuite) DistributeAllRewards() sdk.Coins {
	// We must create at least one lock, otherwise distribution won't work
	lockOwner := apptesting.CreateRandomAccounts(1)[0]
	suite.LockTokens(lockOwner, sdk.NewCoins(sdk.NewInt64Coin("stake", 100)))

	err := suite.App.StreamerKeeper.BeforeEpochStart(suite.Ctx, "day")
	suite.Require().NoError(err)
	coins, err := suite.App.StreamerKeeper.AfterEpochEnd(suite.Ctx, "day")
	suite.Require().NoError(err)
	return coins
}

// LockTokens locks tokens for the specified duration
func (suite *KeeperTestSuite) LockTokens(addr sdk.AccAddress, coins sdk.Coins) {
	suite.FundAcc(addr, coins)
	_, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr, coins, time.Hour)
	suite.Require().NoError(err)
}
