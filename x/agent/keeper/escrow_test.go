package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

const (
	spendDenom  = "adym"
	windowLimit = 1000
	windowLen   = 10
)

// EscrowTestSuite runs against the real app (bank keeper, module account) but
// swaps in a fakeVerifier by building a second agent keeper over the same
// store, since a valid GCP attestation token cannot be produced locally.
type EscrowTestSuite struct {
	apptesting.KeeperTestHelper

	k         *keeper.Keeper
	msgServer types.MsgServer
	verifier  *fakeVerifier
}

func TestEscrowTestSuite(t *testing.T) {
	suite.Run(t, new(EscrowTestSuite))
}

func (s *EscrowTestSuite) SetupTest() {
	s.App = apptesting.Setup(s.T())
	s.Ctx = s.App.NewContext(false)
	s.verifier = &fakeVerifier{}
	key := s.App.GetKVStoreKeys()[types.StoreKey]
	s.k = keeper.NewKeeper(
		s.App.AppCodec(),
		runtime.NewKVStoreService(key),
		s.verifier,
		s.App.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	s.msgServer = keeper.NewMsgServerImpl(*s.k)
}

func (s *EscrowTestSuite) registerAgent(id string) sdk.AccAddress {
	_, _, owner := testdata.KeyTestPubAddr()
	fee, err := s.k.AgentRegistrationFee(s.Ctx)
	s.Require().NoError(err)
	s.FundAcc(owner, sdk.NewCoins(fee))
	_, err = s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), id, validPolicyT(s.T(), "q")))
	s.Require().NoError(err)
	return owner
}

// spendingAgent registers an agent and enables its spend policy.
func (s *EscrowTestSuite) spendingAgent(id string) sdk.AccAddress {
	owner := s.registerAgent(id)
	_, err := s.msgServer.UpdateAgentSpendPolicy(s.Ctx,
		types.NewMsgUpdateAgentSpendPolicy(owner.String(), id, spendDenom, math.NewInt(windowLimit), windowLen))
	s.Require().NoError(err)
	return owner
}

func (s *EscrowTestSuite) fundEscrow(id string, amt int64) {
	_, _, funder := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewCoin(spendDenom, math.NewInt(amt)))
	s.FundAcc(funder, coins)
	_, err := s.msgServer.FundAgentEscrow(s.Ctx, types.NewMsgFundAgentEscrow(funder.String(), id, coins))
	s.Require().NoError(err)
}

// transferMsg builds a transfer whose token equals the nonce the handler
// derives, so the fake verifier accepts it.
func (s *EscrowTestSuite) transferMsg(id string, recipient sdk.AccAddress, amt int64, seq uint64) *types.MsgSubmitAttestedTransfer {
	_, _, sub := testdata.KeyTestPubAddr()
	amount := math.NewInt(amt)
	memo := []byte("m")
	payload := types.AttestedTransferBytes(recipient.String(), spendDenom, amount, memo)
	return &types.MsgSubmitAttestedTransfer{
		Submitter: sub.String(),
		AgentId:   id,
		Recipient: recipient.String(),
		Amount:    amount,
		Memo:      memo,
		Token:     types.TransferNonce(id, payload, seq),
	}
}

func (s *EscrowTestSuite) moduleBalance() math.Int {
	return s.App.BankKeeper.GetBalance(s.Ctx, authtypes.NewModuleAddress(types.ModuleName), spendDenom).Amount
}

func (s *EscrowTestSuite) TestFundTransfer_HappyPath() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)

	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(500), s.moduleBalance())

	_, _, recipient := testdata.KeyTestPubAddr()
	res, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 200, 0))
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), res.Seq)

	// recipient credited exactly; ledger and module account both debited
	s.Require().Equal(math.NewInt(200), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
	s.Require().Equal(math.NewInt(300), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(300), s.moduleBalance())

	// transfer is in the action log with the nonce-committed payload
	entry, found := s.k.GetActionLogEntry(s.Ctx, "a1", 0)
	s.Require().True(found)
	s.Require().Equal(types.AttestedTransferBytes(recipient.String(), spendDenom, math.NewInt(200), []byte("m")), entry.Payload)

	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(1), agent.ActionSeq)
	s.Require().Equal(math.NewInt(200), agent.SpendWindowSpent)
}

func (s *EscrowTestSuite) TestTransfer_RevokedPolicyRejected() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)

	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	fp, err := types.PolicyFingerprint(agent.Policy)
	s.Require().NoError(err)
	s.Require().NoError(s.k.SetRevoked(s.Ctx, fp))

	_, _, recipient := testdata.KeyTestPubAddr()
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 200, 0))
	s.Require().ErrorContains(err, "policy revoked")
	s.Require().Equal(math.ZeroInt(), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))

	agent, found = s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Zero(agent.ActionSeq)
}

func (s *EscrowTestSuite) TestTransfer_ExceedsWindowCap_NothingChanges() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 5000)
	_, _, recipient := testdata.KeyTestPubAddr()

	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, windowLimit+1, 0))
	s.Require().ErrorIs(err, types.ErrSpendBudgetExceeded)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).IsZero())
	s.Require().Equal(math.NewInt(5000), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(5000), s.moduleBalance())
	_, found := s.k.GetActionLogEntry(s.Ctx, "a1", 0)
	s.Require().False(found)
	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(0), agent.ActionSeq)
}

func (s *EscrowTestSuite) TestTransfer_WindowRollsOver() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 5000)
	_, _, recipient := testdata.KeyTestPubAddr()

	// exhaust the window's budget mid-bucket
	s.Ctx = s.Ctx.WithBlockHeight(15)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, windowLimit, 0))
	s.Require().NoError(err)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 1, 1))
	s.Require().ErrorIs(err, types.ErrSpendBudgetExceeded)

	// same bucket (heights 10..19): still exhausted
	s.Ctx = s.Ctx.WithBlockHeight(19)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 1, 1))
	s.Require().ErrorIs(err, types.ErrSpendBudgetExceeded)

	// next bucket: full budget available again
	s.Ctx = s.Ctx.WithBlockHeight(20)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, windowLimit, 1))
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(2*windowLimit), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
}

func (s *EscrowTestSuite) TestTransfer_NonceBindsRecipientAndAmount() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	// token minted for a different recipient
	msg := s.transferMsg("a1", recipient, 100, 0)
	_, _, other := testdata.KeyTestPubAddr()
	msg.Recipient = other.String()
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().ErrorContains(err, "verify attestation")

	// token minted for a different amount
	msg = s.transferMsg("a1", recipient, 100, 0)
	msg.Amount = math.NewInt(400)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().ErrorContains(err, "verify attestation")

	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
}

// TestCrossDomainReplayRejected pins the nonce domain separation: a token
// minted for a pending transfer cannot be replayed as a plain action at the
// same seq (which would advance the counter and kill the enclave-authorized
// payment), and vice versa.
func (s *EscrowTestSuite) TestCrossDomainReplayRejected() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	// observer copies a pending transfer's payload+token and submits it as an action
	transfer := s.transferMsg("a1", recipient, 100, 0)
	payload := types.AttestedTransferBytes(recipient.String(), spendDenom, transfer.Amount, transfer.Memo)
	_, err := s.msgServer.SubmitAttestedAction(s.Ctx, &types.MsgSubmitAttestedAction{
		Submitter: transfer.Submitter,
		AgentId:   "a1",
		Payload:   payload,
		Token:     transfer.Token,
	})
	s.Require().ErrorContains(err, "verify attestation")

	// seq unchanged: the original transfer still goes through
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, transfer)
	s.Require().NoError(err)

	// and an action token cannot authorize a transfer of the same payload
	actionToken := types.ActionNonce("a1", payload, 1)
	transfer2 := s.transferMsg("a1", recipient, 100, 1)
	transfer2.Token = actionToken
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, transfer2)
	s.Require().ErrorContains(err, "verify attestation")
}

func (s *EscrowTestSuite) TestTransfer_PayloadTooLarge() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	params, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)

	msg := s.transferMsg("a1", recipient, 100, 0)
	msg.Memo = make([]byte, params.MaxActionBytes) // payload = memo + transfer fields > max
	payload := types.AttestedTransferBytes(recipient.String(), spendDenom, msg.Amount, msg.Memo)
	msg.Token = types.TransferNonce("a1", payload, 0)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().ErrorContains(err, "max action bytes")

	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(0), agent.ActionSeq)
}

func (s *EscrowTestSuite) TestTransfer_ReplayRejected() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	msg := s.transferMsg("a1", recipient, 100, 0)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().NoError(err)

	// same (transfer, token) again: seq advanced, nonce differs, verifier rejects
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().ErrorContains(err, "verify attestation")
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
}

func (s *EscrowTestSuite) TestTransfer_InsufficientEscrow_NoPartialPayout() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 50)
	_, _, recipient := testdata.KeyTestPubAddr()

	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 80, 0))
	s.Require().ErrorIs(err, types.ErrInsufficientEscrow)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).IsZero())
	s.Require().Equal(math.NewInt(50), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(0), agent.ActionSeq)
}

func (s *EscrowTestSuite) TestTransfer_SpendingDisabled() {
	s.registerAgent("a1") // no spend policy
	_, _, recipient := testdata.KeyTestPubAddr()

	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 1, 0))
	s.Require().ErrorIs(err, types.ErrSpendingDisabled)

	// the same agent can still submit plain attested actions
	res, err := s.msgServer.SubmitAttestedAction(s.Ctx, validMsg(s.T(), "a1", []byte("p"), 0))
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), res.Seq)
}

func (s *EscrowTestSuite) TestTransfer_SharesActionLogWithActions() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	_, err := s.msgServer.SubmitAttestedAction(s.Ctx, validMsg(s.T(), "a1", []byte("p0"), 0))
	s.Require().NoError(err)

	res, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 1))
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), res.Seq)

	_, err = s.msgServer.SubmitAttestedAction(s.Ctx, validMsg(s.T(), "a1", []byte("p2"), 2))
	s.Require().NoError(err)

	for seq := uint64(0); seq < 3; seq++ {
		_, found := s.k.GetActionLogEntry(s.Ctx, "a1", seq)
		s.Require().True(found, "seq %d", seq)
	}
	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(3), agent.ActionSeq)
}

func (s *EscrowTestSuite) TestFund_UnknownAgentRejected() {
	_, _, funder := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewCoin(spendDenom, math.NewInt(10)))
	s.FundAcc(funder, coins)

	_, err := s.msgServer.FundAgentEscrow(s.Ctx, types.NewMsgFundAgentEscrow(funder.String(), "ghost", coins))
	s.Require().ErrorIs(err, types.ErrAgentNotFound)
}

func (s *EscrowTestSuite) TestWithdraw_OwnerOnlyAndNoUnderflow() {
	owner := s.spendingAgent("a1")
	s.fundEscrow("a1", 300)
	coins := func(amt int64) sdk.Coins { return sdk.NewCoins(sdk.NewCoin(spendDenom, math.NewInt(amt))) }

	// non-owner rejected
	_, _, other := testdata.KeyTestPubAddr()
	_, err := s.msgServer.WithdrawAgentEscrow(s.Ctx, types.NewMsgWithdrawAgentEscrow(other.String(), "a1", coins(100)))
	s.Require().ErrorIs(err, types.ErrUnauthorized)

	// underflow rejected
	_, err = s.msgServer.WithdrawAgentEscrow(s.Ctx, types.NewMsgWithdrawAgentEscrow(owner.String(), "a1", coins(301)))
	s.Require().ErrorIs(err, types.ErrInsufficientEscrow)

	// owner withdraws
	before := s.App.BankKeeper.GetBalance(s.Ctx, owner, spendDenom).Amount
	_, err = s.msgServer.WithdrawAgentEscrow(s.Ctx, types.NewMsgWithdrawAgentEscrow(owner.String(), "a1", coins(300)))
	s.Require().NoError(err)
	s.Require().Equal(before.AddRaw(300), s.App.BankKeeper.GetBalance(s.Ctx, owner, spendDenom).Amount)
	s.Require().True(s.k.GetEscrowBalance(s.Ctx, "a1").IsZero())
	s.Require().True(s.moduleBalance().IsZero())
}

func (s *EscrowTestSuite) TestUpdateSpendPolicy_OwnerOnly() {
	s.spendingAgent("a1")

	_, _, other := testdata.KeyTestPubAddr()
	_, err := s.msgServer.UpdateAgentSpendPolicy(s.Ctx,
		types.NewMsgUpdateAgentSpendPolicy(other.String(), "a1", spendDenom, math.NewInt(1), 1))
	s.Require().ErrorIs(err, types.ErrUnauthorized)
}

func (s *EscrowTestSuite) TestUpdateSpendPolicy_RecipientAllowlistPersistedQueriedAndEmitted() {
	owner := s.registerAgent("a1")
	_, _, recipient := testdata.KeyTestPubAddr()
	msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", spendDenom, math.NewInt(windowLimit), windowLen)
	msg.RecipientAllowlist = []string{recipient.String()}

	_, err := s.msgServer.UpdateAgentSpendPolicy(s.Ctx, msg)
	s.Require().NoError(err)

	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Equal(msg.RecipientAllowlist, agent.SpendRecipientAllowlist)
	query, err := s.k.Agent(s.Ctx, &types.QueryAgentRequest{AgentId: "a1"})
	s.Require().NoError(err)
	s.Require().Equal(msg.RecipientAllowlist, query.Agent.SpendRecipientAllowlist)

	events := s.Ctx.EventManager().Events()
	s.Require().NotEmpty(events)
	last := events[len(events)-1]
	s.Require().Equal("dymensionxyz.dymension.agent.EventUpdateAgentSpendPolicy", last.Type)
	foundAllowlist := false
	for _, attr := range last.Attributes {
		if attr.Key == "recipient_allowlist" && attr.Value == `["`+recipient.String()+`"]` {
			foundAllowlist = true
		}
	}
	s.Require().True(foundAllowlist)
}

func (s *EscrowTestSuite) TestUpdateSpendPolicy_RecipientAllowlistAboveMaxRejected() {
	owner := s.registerAgent("a1")
	params, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	params.SpendRecipientAllowlistMax = 1
	s.Require().NoError(s.k.SetParams(s.Ctx, params))
	_, _, first := testdata.KeyTestPubAddr()
	_, _, second := testdata.KeyTestPubAddr()
	msg := types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", spendDenom, math.NewInt(windowLimit), windowLen)
	msg.RecipientAllowlist = []string{first.String(), second.String()}

	_, err = s.msgServer.UpdateAgentSpendPolicy(s.Ctx, msg)
	s.Require().Error(err)
	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Empty(agent.SpendRecipientAllowlist)
}

func (s *EscrowTestSuite) TestMigrate1to2_DefaultsRecipientAllowlistMax() {
	params := types.DefaultParams()
	params.SpendRecipientAllowlistMax = 0
	s.Require().NoError(s.k.SetParams(s.Ctx, params))

	migrator := keeper.NewMigrator(*s.k)
	s.Require().NoError(migrator.Migrate1to2(s.Ctx))

	params, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(types.DefaultSpendRecipientAllowlistMax), params.SpendRecipientAllowlistMax)
}

func (s *EscrowTestSuite) TestTransfer_RecipientAllowlist() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, allowed := testdata.KeyTestPubAddr()
	_, _, denied := testdata.KeyTestPubAddr()
	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	agent.SpendRecipientAllowlist = []string{allowed.String()}
	s.Require().NoError(s.k.SetAgent(s.Ctx, agent))

	eventsBefore := len(s.Ctx.EventManager().Events())
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", denied, 100, 0))
	s.Require().ErrorIs(err, types.ErrRecipientNotAllowed)
	s.Require().Equal(eventsBefore, len(s.Ctx.EventManager().Events()))
	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, denied, spendDenom).IsZero())
	_, found = s.k.GetActionLogEntry(s.Ctx, "a1", 0)
	s.Require().False(found)
	agent, found = s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Zero(agent.ActionSeq)
	s.Require().True(agent.SpendWindowSpent.IsZero())
	s.Require().Zero(s.verifier.calls)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", allowed, 100, 0))
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, allowed, spendDenom).Amount)
}

// TestEscrowSolvencyInvariant exercises the invariant across the full
// lifecycle: fund, transfer, withdraw.
func (s *EscrowTestSuite) TestEscrowSolvencyInvariant() {
	inv := keeper.AllInvariants(*s.k)
	check := func() {
		s.T().Helper()
		msg, broken := inv(s.Ctx)
		s.Require().False(broken, msg)
	}

	check()
	owner := s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	check()

	_, _, recipient := testdata.KeyTestPubAddr()
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 200, 0))
	s.Require().NoError(err)
	check()

	_, err = s.msgServer.WithdrawAgentEscrow(s.Ctx,
		types.NewMsgWithdrawAgentEscrow(owner.String(), "a1", sdk.NewCoins(sdk.NewCoin(spendDenom, math.NewInt(300)))))
	s.Require().NoError(err)
	check()
}

func (s *EscrowTestSuite) TestEscrowBalanceQuery() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	s.Ctx = s.Ctx.WithBlockHeight(15)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 200, 0))
	s.Require().NoError(err)

	res, err := s.k.EscrowBalance(s.Ctx, &types.QueryEscrowBalanceRequest{AgentId: "a1"})
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(300), res.Balance.AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(windowLimit-200), res.RemainingWindowBudget)

	// next window: full budget again
	s.Ctx = s.Ctx.WithBlockHeight(20)
	res, err = s.k.EscrowBalance(s.Ctx, &types.QueryEscrowBalanceRequest{AgentId: "a1"})
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(windowLimit), res.RemainingWindowBudget)
}
