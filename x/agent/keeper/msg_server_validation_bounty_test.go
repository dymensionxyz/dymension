package keeper_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

const (
	bountyDenom    = "adym"
	bountyAmount   = 5_000
	bountyExpiry   = 50
	bountyReqStart = 10
)

// ValidationBountyTestSuite uses the real bank keeper and module account, with
// a fakeVerifier so attested verdicts can be produced locally.
type ValidationBountyTestSuite struct {
	apptesting.KeeperTestHelper

	k              *keeper.Keeper
	msgServer      types.MsgServer
	fee            sdk.Coin
	requester      sdk.AccAddress
	validatorOwner string
	hash           []byte
}

func TestValidationBountyTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationBountyTestSuite))
}

func (s *ValidationBountyTestSuite) SetupTest() {
	s.App = apptesting.Setup(s.T())
	s.Ctx = s.App.NewContext(false).WithBlockHeight(bountyReqStart)
	key := s.App.GetKVStoreKeys()[types.StoreKey]
	s.k = keeper.NewKeeper(s.App.AppCodec(), runtime.NewKVStoreService(key), &fakeVerifier{}, s.App.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	s.msgServer = keeper.NewMsgServerImpl(*s.k)

	p, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	p.ValidationBountyExpiryBlocks = bountyExpiry
	s.Require().NoError(s.k.SetParams(s.Ctx, p))
	s.fee = p.ValidationRequestFee
	s.Require().True(s.fee.IsPositive())

	_, _, validatorOwner := testdata.KeyTestPubAddr()
	_, _, subjectOwner := testdata.KeyTestPubAddr()
	_, _, requester := testdata.KeyTestPubAddr()
	s.validatorOwner, s.requester = validatorOwner.String(), requester
	s.Require().NoError(s.k.SetAgent(s.Ctx, types.Agent{Id: "validator", Owner: s.validatorOwner, Active: true, Policy: policyA(), ActionSeq: 3}))
	s.Require().NoError(s.k.SetAgent(s.Ctx, types.Agent{Id: "subject", Owner: subjectOwner.String(), Active: true, Policy: policyB(), ActionSeq: 2}))
	s.hash = bytes.Repeat([]byte{1}, 32)
	s.FundAcc(requester, sdk.NewCoins(sdk.NewCoin(bountyDenom, s.fee.Amount.MulRaw(2).AddRaw(bountyAmount*2))))
}

func (s *ValidationBountyTestSuite) bounty() sdk.Coin {
	return sdk.NewInt64Coin(bountyDenom, bountyAmount)
}

func (s *ValidationBountyTestSuite) request(bounty sdk.Coin) error {
	msg := types.NewMsgRequestValidation(s.requester.String(), "validator", "subject", 1, s.hash, "")
	msg.Bounty = bounty
	_, err := s.msgServer.RequestValidation(s.Ctx, msg)
	return err
}

// respondAttested builds a verdict whose token equals the nonce the handler
// derives for the validator's current action sequence.
func (s *ValidationBountyTestSuite) respondAttested() error {
	validator, found := s.k.GetAgent(s.Ctx, "validator")
	s.Require().True(found)
	msg := &types.MsgRespondValidationAttested{Responder: s.validatorOwner, RequestHash: s.hash, Response: 100, Tag: "tee"}
	payload := types.AttestedValidationBytes(msg.RequestHash, msg.Response, msg.ResponseHash, msg.ResponseUri, msg.Tag, "subject", 1)
	msg.Token = types.ValidationNonce("validator", payload, validator.ActionSeq)
	_, err := s.msgServer.RespondValidationAttested(s.Ctx, msg)
	return err
}

func (s *ValidationBountyTestSuite) respondPlain() error {
	_, err := s.msgServer.RespondValidation(s.Ctx, types.NewMsgRespondValidation(s.validatorOwner, s.hash, 40, "", nil, "plain"))
	return err
}

func (s *ValidationBountyTestSuite) reclaim(requester string) error {
	_, err := s.msgServer.ReclaimValidationBounty(s.Ctx, types.NewMsgReclaimValidationBounty(requester, s.hash))
	return err
}

func (s *ValidationBountyTestSuite) requesterBalance() math.Int {
	return s.App.BankKeeper.GetBalance(s.Ctx, s.requester, bountyDenom).Amount
}

func (s *ValidationBountyTestSuite) moduleBalance() math.Int {
	return s.App.BankKeeper.GetBalance(s.Ctx, authtypes.NewModuleAddress(types.ModuleName), bountyDenom).Amount
}

func (s *ValidationBountyTestSuite) validatorEscrow() math.Int {
	return s.k.GetEscrowBalance(s.Ctx, "validator").AmountOf(bountyDenom)
}

func (s *ValidationBountyTestSuite) requireInt(expected, actual math.Int) {
	s.T().Helper()
	s.Require().True(expected.Equal(actual), "expected %s, got %s", expected, actual)
}

func (s *ValidationBountyTestSuite) storedRequest() types.ValidationRequest {
	req, found := s.k.GetValidationRequest(s.Ctx, s.hash)
	s.Require().True(found)
	return req
}

func (s *ValidationBountyTestSuite) countEvents(eventType string) int {
	n := 0
	for _, e := range s.Ctx.EventManager().Events() {
		if e.Type == eventType {
			n++
		}
	}
	return n
}

func (s *ValidationBountyTestSuite) TestRequestEscrowsBountyAndBurnsFee() {
	balance, module := s.requesterBalance(), s.moduleBalance()
	supply := s.App.BankKeeper.GetSupply(s.Ctx, bountyDenom).Amount

	s.Require().NoError(s.request(s.bounty()))

	s.requireInt(balance.Sub(s.fee.Amount).SubRaw(bountyAmount), s.requesterBalance())
	s.requireInt(module.AddRaw(bountyAmount), s.moduleBalance())
	s.requireInt(supply.Sub(s.fee.Amount), s.App.BankKeeper.GetSupply(s.Ctx, bountyDenom).Amount)
	req := s.storedRequest()
	s.Require().Equal(s.bounty(), req.Bounty)
	s.Require().False(req.BountyClaimed)
	s.Require().Zero(s.validatorEscrow().Int64())
	s.Require().Equal(1, s.countEvents("dymensionxyz.dymension.agent.EventRequestValidation"))
}

func (s *ValidationBountyTestSuite) TestRequestRejectsUnaffordableBounty() {
	err := s.request(sdk.NewCoin(bountyDenom, s.requesterBalance()))
	s.Require().Error(err)
	_, found := s.k.GetValidationRequest(s.Ctx, s.hash)
	s.Require().False(found)
}

func (s *ValidationBountyTestSuite) TestFirstAttestedVerdictClaimsBountyOnce() {
	s.Require().NoError(s.request(s.bounty()))
	module := s.moduleBalance()

	s.Require().NoError(s.respondAttested())
	s.Require().Equal(int64(bountyAmount), s.validatorEscrow().Int64())
	s.Require().True(s.storedRequest().BountyClaimed)
	s.Require().Equal(1, s.countEvents("dymensionxyz.dymension.agent.EventClaimValidationBounty"))
	s.requireInt(module, s.moduleBalance())

	s.Require().NoError(s.respondAttested())
	s.Require().Equal(int64(bountyAmount), s.validatorEscrow().Int64())
	s.Require().Equal(1, s.countEvents("dymensionxyz.dymension.agent.EventClaimValidationBounty"))
	s.Require().Equal(uint64(2), s.storedRequest().ResponseCount)

	s.Ctx = s.Ctx.WithBlockHeight(bountyReqStart + bountyExpiry)
	s.Require().ErrorIs(s.reclaim(s.requester.String()), gerrc.ErrInvalidArgument)

	_, err := s.msgServer.WithdrawAgentEscrow(s.Ctx, types.NewMsgWithdrawAgentEscrow(s.validatorOwner, "validator", sdk.NewCoins(s.bounty())))
	s.Require().NoError(err)
	s.requireInt(module.SubRaw(bountyAmount), s.moduleBalance())

	s.Require().NoError(keeper.InvariantEscrowSolvency(*s.k)(s.Ctx))
}

func (s *ValidationBountyTestSuite) TestPlainVerdictDoesNotClaimBounty() {
	s.Require().NoError(s.request(s.bounty()))

	s.Require().NoError(s.respondPlain())
	s.Require().Zero(s.validatorEscrow().Int64())
	s.Require().False(s.storedRequest().BountyClaimed)
	s.Require().Zero(s.countEvents("dymensionxyz.dymension.agent.EventClaimValidationBounty"))

	s.Require().NoError(s.respondAttested())
	s.Require().Equal(int64(bountyAmount), s.validatorEscrow().Int64())
}

func (s *ValidationBountyTestSuite) TestReclaimAfterExpiry() {
	s.Require().NoError(s.request(s.bounty()))
	balance, module := s.requesterBalance(), s.moduleBalance()

	s.Ctx = s.Ctx.WithBlockHeight(bountyReqStart + bountyExpiry - 1)
	err := s.reclaim(s.requester.String())
	s.Require().ErrorIs(err, gerrc.ErrInvalidArgument)
	s.Require().ErrorContains(err, "bounty not yet expired")

	s.Ctx = s.Ctx.WithBlockHeight(bountyReqStart + bountyExpiry)
	_, _, stranger := testdata.KeyTestPubAddr()
	s.Require().ErrorIs(s.reclaim(stranger.String()), types.ErrUnauthorized)
	s.requireInt(balance, s.requesterBalance())

	s.Require().NoError(s.reclaim(s.requester.String()))
	s.requireInt(balance.AddRaw(bountyAmount), s.requesterBalance())
	s.requireInt(module.SubRaw(bountyAmount), s.moduleBalance())
	req := s.storedRequest()
	s.Require().True(req.BountyClaimed)
	s.Require().True(req.Bounty.IsZero())
	s.Require().Equal(1, s.countEvents("dymensionxyz.dymension.agent.EventReclaimValidationBounty"))

	err = s.reclaim(s.requester.String())
	s.Require().ErrorContains(err, "no reclaimable bounty")
	s.requireInt(balance.AddRaw(bountyAmount), s.requesterBalance())

	s.Require().NoError(s.respondAttested())
	s.Require().Zero(s.validatorEscrow().Int64())
}

func (s *ValidationBountyTestSuite) TestNoBountyKeepsCurrentBehavior() {
	for name, bounty := range map[string]sdk.Coin{"unset": {}, "zero": sdk.NewInt64Coin(bountyDenom, 0)} {
		s.Run(name, func() {
			s.SetupTest()
			balance, module := s.requesterBalance(), s.moduleBalance()

			s.Require().NoError(s.request(bounty))
			s.requireInt(balance.Sub(s.fee.Amount), s.requesterBalance())
			s.requireInt(module, s.moduleBalance())
			s.Require().False(s.storedRequest().BountyClaimed)

			s.Require().NoError(s.respondAttested())
			s.Require().Zero(s.validatorEscrow().Int64())
			s.Require().Zero(s.countEvents("dymensionxyz.dymension.agent.EventClaimValidationBounty"))

			s.Ctx = s.Ctx.WithBlockHeight(bountyReqStart + bountyExpiry)
			s.Require().ErrorContains(s.reclaim(s.requester.String()), "no reclaimable bounty")
		})
	}
}

func (s *ValidationBountyTestSuite) TestReclaimMissingRequest() {
	s.Require().ErrorIs(s.reclaim(s.requester.String()), types.ErrValidationRequestNotFound)
}

func (s *ValidationBountyTestSuite) TestGenesisRoundTripsBounty() {
	s.Require().NoError(s.request(s.bounty()))
	claimedHash := s.hash
	s.Require().NoError(s.respondAttested())
	s.hash = bytes.Repeat([]byte{2}, 32)
	s.Require().NoError(s.request(s.bounty()))

	exported := keeper.ExportGenesis(s.Ctx, s.k)
	s.Require().NoError(exported.Validate())
	cdc := codec.NewProtoCodec(s.App.InterfaceRegistry())
	bz, err := cdc.MarshalJSON(exported)
	s.Require().NoError(err)
	var imported types.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(bz, &imported))

	ctx, k, _ := setup(s.T())
	keeper.InitGenesis(ctx, k, imported)
	claimed, found := k.GetValidationRequest(ctx, claimedHash)
	s.Require().True(found)
	s.Require().Equal(s.bounty(), claimed.Bounty)
	s.Require().True(claimed.BountyClaimed)
	open, found := k.GetValidationRequest(ctx, s.hash)
	s.Require().True(found)
	s.Require().Equal(s.bounty(), open.Bounty)
	s.Require().False(open.BountyClaimed)
	p, err := k.GetParams(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(bountyExpiry), p.ValidationBountyExpiryBlocks)
}

func (s *ValidationBountyTestSuite) TestMigrate3to4_DefaultsBountyExpiry() {
	p, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	p.ValidationBountyExpiryBlocks = 0
	s.Require().NoError(s.k.SetParams(s.Ctx, p))

	s.Require().NoError(keeper.NewMigrator(*s.k).Migrate3to4(s.Ctx))

	p, err = s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(types.DefaultValidationBountyExpiryBlocks), p.ValidationBountyExpiryBlocks)
}
