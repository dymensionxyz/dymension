package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

const expiryHeight = 100

// setSpendExpiry re-issues the default spend policy with the given expiry, as
// an owner renewing the mandate would.
func (s *EscrowTestSuite) setSpendExpiry(owner sdk.AccAddress, id string, expiry uint64) {
	_, err := s.msgServer.UpdateAgentSpendPolicy(s.Ctx,
		types.NewMsgUpdateAgentSpendPolicy(owner.String(), id, spendDenom, math.NewInt(windowLimit), windowLen, expiry))
	s.Require().NoError(err)
}

func (s *EscrowTestSuite) TestSpendExpiry_ZeroNeverExpires() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	s.Ctx = s.Ctx.WithBlockHeight(1_000_000)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 0))
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
}

func (s *EscrowTestSuite) TestSpendExpiry_AtBoundaryAllowed() {
	owner := s.spendingAgent("a1")
	s.setSpendExpiry(owner, "a1", expiryHeight)
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	s.Ctx = s.Ctx.WithBlockHeight(expiryHeight)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 0))
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
}

func (s *EscrowTestSuite) TestSpendExpiry_PastExpiryRejected_NothingChanges() {
	owner := s.spendingAgent("a1")
	s.setSpendExpiry(owner, "a1", expiryHeight)
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	before, _ := s.k.GetAgent(s.Ctx, "a1")

	s.Ctx = s.Ctx.WithBlockHeight(expiryHeight + 1)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 0))
	s.Require().ErrorIs(err, types.ErrSpendAuthorityExpired)

	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).IsZero())
	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(500), s.moduleBalance())
	_, found := s.k.GetActionLogEntry(s.Ctx, "a1", 0)
	s.Require().False(found)
	after, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(before, after)
}

func (s *EscrowTestSuite) TestSpendExpiry_RenewalRestoresAuthority() {
	owner := s.spendingAgent("a1")
	s.setSpendExpiry(owner, "a1", expiryHeight)
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	s.Ctx = s.Ctx.WithBlockHeight(expiryHeight + 1)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 0))
	s.Require().ErrorIs(err, types.ErrSpendAuthorityExpired)

	s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
	s.setSpendExpiry(owner, "a1", 2*expiryHeight)
	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Equal(uint64(2*expiryHeight), agent.SpendExpiryHeight)
	s.requireSpendPolicyEventExpiry("200")

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, 0))
	s.Require().NoError(err)
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
}

func (s *EscrowTestSuite) TestSpendExpiry_DisabledWithExpiryRejected() {
	owner := s.registerAgent("a1")

	_, err := s.msgServer.UpdateAgentSpendPolicy(s.Ctx,
		types.NewMsgUpdateAgentSpendPolicy(owner.String(), "a1", "", math.ZeroInt(), 0, expiryHeight))
	s.Require().ErrorContains(err, "spend expiry set without spend denom")
	agent, found := s.k.GetAgent(s.Ctx, "a1")
	s.Require().True(found)
	s.Require().Zero(agent.SpendExpiryHeight)
}

func (s *EscrowTestSuite) TestSpendExpiry_DoesNotGateAttestedActions() {
	owner := s.spendingAgent("a1")
	s.setSpendExpiry(owner, "a1", expiryHeight)

	s.Ctx = s.Ctx.WithBlockHeight(expiryHeight + 1)
	res, err := s.msgServer.SubmitAttestedAction(s.Ctx, validMsg(s.T(), "a1", []byte("p"), 0))
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), res.Seq)
}

func (s *EscrowTestSuite) requireSpendPolicyEventExpiry(want string) {
	events := s.Ctx.EventManager().Events()
	s.Require().NotEmpty(events)
	last := events[len(events)-1]
	s.Require().Equal("dymensionxyz.dymension.agent.EventUpdateAgentSpendPolicy", last.Type)
	for _, attr := range last.Attributes {
		if attr.Key == "spend_expiry_height" {
			s.Require().Equal(want, attr.Value)
			return
		}
	}
	s.Fail("spend_expiry_height attribute missing")
}
