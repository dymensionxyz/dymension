package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
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
	app := apptesting.Setup(s.T())
	ctx := app.NewContext(false)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.SponsorshipKeeper))
	queryClient := types.NewQueryClient(queryHelper)
	msgServer := keeper.NewMsgServer(app.SponsorshipKeeper)

	s.App = app
	s.Ctx = ctx
	s.queryClient = queryClient
	s.msgServer = msgServer

	s.SetDefaultTestParams()
}

func (s *KeeperTestSuite) CreateAssetGauge() uint64 {
	s.T().Helper()

	gaugeID, err := s.App.IncentivesKeeper.CreateAssetGauge(
		s.Ctx,
		true,
		s.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		lockuptypes.QueryCondition{
			Denom:    "stake",
			Duration: time.Hour,
		},
		s.Ctx.BlockTime(),
		1,
	)
	s.Require().NoError(err)
	return gaugeID
}

func (s *KeeperTestSuite) CreateEndorsementGauge(rollappId string) uint64 {
	s.T().Helper()

	gaugeID, err := s.App.IncentivesKeeper.CreateEndorsementGauge(
		s.Ctx,
		true,
		s.App.AccountKeeper.GetModuleAddress(types.ModuleName),
		sdk.Coins{},
		incentivestypes.EndorsementGauge{RollappId: rollappId},
		time.Now(),
		1,
	)
	s.Require().NoError(err)
	return gaugeID
}

func (s *KeeperTestSuite) CreateGauges(num int) {
	s.T().Helper()

	for i := 0; i < num; i++ {
		s.CreateAssetGauge()
	}
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

func (s *KeeperTestSuite) RevokeVote(vote types.MsgRevokeVote) {
	s.T().Helper()

	resp, err := s.msgServer.RevokeVote(s.Ctx, &vote)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}

func (s *KeeperTestSuite) CreateValidator() stakingtypes.ValidatorI {
	s.T().Helper()

	valAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, types.DYM.MulRaw(1_000))

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
	s.Require().NoError(err)

	// Create a validator
	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	resp, err := stakingMsgSrv.CreateValidator(s.Ctx, msgCreate)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	val, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	return val
}

func (s *KeeperTestSuite) CreateValidatorWithAddress(acc sdk.AccAddress, balance math.Int) stakingtypes.ValidatorI {
	s.T().Helper()

	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)

	initCoin := sdk.NewCoin(bondDenom, balance)
	initCoins := sdk.NewCoins(initCoin)
	apptesting.FundAccount(s.App, s.Ctx, acc, initCoins)

	// Build MsgCreateValidator
	valAddr := sdk.ValAddress(acc)
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr.String(),
		privEd.PubKey(),
		initCoin,
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
		math.OneInt(),
	)
	s.Require().NoError(err)

	// Create a validator
	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	resp, err := stakingMsgSrv.CreateValidator(s.Ctx, msgCreate)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	val, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	return val
}

func (s *KeeperTestSuite) CreateDelegator(valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	delAddrs := apptesting.AddTestAddrs(s.App, s.Ctx, 1, types.DYM.MulRaw(1_000))
	delAddr := delAddrs[0]
	return s.Delegate(delAddr, valAddr, coin)
}

func (s *KeeperTestSuite) Delegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	resp, err := stakingMsgSrv.Delegate(s.Ctx, stakingtypes.NewMsgDelegate(delAddr.String(), valAddr.String(), coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	del, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delAddr, valAddr)
	s.Require().NoError(err)

	return del
}

// Undelegate sends MsgUndelegate and returns the delegation object. Return value might me nil in case if
// the delegator completely unbonds.
func (s *KeeperTestSuite) Undelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, coin sdk.Coin) stakingtypes.DelegationI {
	s.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	_, err := stakingMsgSrv.Undelegate(s.Ctx, stakingtypes.NewMsgUndelegate(delAddr.String(), valAddr.String(), coin))
	s.Require().NoError(err)

	del, _ := s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valAddr)
	return del
}

// Undelegate sends MsgUndelegate and returns the delegation object. Src return value might me nil in case if
// the delegator completely unbonds.
func (s *KeeperTestSuite) BeginRedelegate(
	delAddr sdk.AccAddress,
	valSrcAddr, valDstAddr sdk.ValAddress,
	coin sdk.Coin,
) (src, dst stakingtypes.DelegationI) {
	s.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	resp, err := stakingMsgSrv.BeginRedelegate(s.Ctx, stakingtypes.NewMsgBeginRedelegate(delAddr.String(), valSrcAddr.String(), valDstAddr.String(), coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	src, _ = s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valSrcAddr)
	dst, _ = s.App.StakingKeeper.Delegation(s.Ctx, delAddr, valDstAddr)
	return src, dst
}

func (s *KeeperTestSuite) CancelUnbondingDelegation(
	delAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	creationHeight int64,
	coin sdk.Coin,
) stakingtypes.DelegationI {
	s.T().Helper()

	stakingMsgSrv := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	resp, err := stakingMsgSrv.CancelUnbondingDelegation(s.Ctx, stakingtypes.NewMsgCancelUnbondingDelegation(delAddr.String(), valAddr.String(), creationHeight, coin))
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	src, err := s.App.StakingKeeper.GetDelegation(s.Ctx, delAddr, valAddr)
	s.Require().NoError(err)

	return src
}

func (s *KeeperTestSuite) AssertHaveDelegatorValidator(voterAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	s.T().Helper()

	have := s.haveDelegatorValidator(voterAddr, valAddr)
	s.Require().True(have)
}

func (s *KeeperTestSuite) AssertNotHaveDelegatorValidator(voterAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	s.T().Helper()

	have := s.haveDelegatorValidator(voterAddr, valAddr)
	s.Require().False(have)
}

func (s *KeeperTestSuite) haveDelegatorValidator(voterAddr sdk.AccAddress, valAddr sdk.ValAddress) bool {
	have, err := s.App.SponsorshipKeeper.HasDelegatorValidatorPower(s.Ctx, voterAddr, valAddr)
	s.Require().NoError(err)
	return have
}

func (s *KeeperTestSuite) AssertVoted(voterAddr sdk.AccAddress) {
	s.T().Helper()

	voted, err := s.App.SponsorshipKeeper.Voted(s.Ctx, voterAddr)
	s.Require().NoError(err)
	s.Require().True(voted)
}

func (s *KeeperTestSuite) AssertNotVoted(voterAddr sdk.AccAddress) {
	s.T().Helper()

	voted, err := s.App.SponsorshipKeeper.Voted(s.Ctx, voterAddr)
	s.Require().NoError(err)
	s.Require().False(voted)
}

func (s *KeeperTestSuite) AssertDelegatorValidator(delAddr sdk.AccAddress, valAddr sdk.ValAddress, expectedPower math.Int) {
	s.T().Helper()

	vp, err := s.App.SponsorshipKeeper.GetDelegatorValidatorPower(s.Ctx, delAddr, valAddr)
	s.Require().NoError(err)
	s.Require().Equal(expectedPower, vp)
}

// SetDefaultTestParams sets module params with MinVotingPower = 1 for convenience.
func (s *KeeperTestSuite) SetDefaultTestParams() {
	err := s.App.SponsorshipKeeper.SetParams(s.Ctx, DefaultTestParams())
	s.Require().NoError(err)
}

// DefaultTestParams returns module params with MinVotingPower = 1 for convenience.
func DefaultTestParams() types.Params {
	params := types.DefaultParams()
	params.MinVotingPower = math.NewInt(1)
	return params
}
