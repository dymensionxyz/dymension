package keeper_test

import (
	"encoding/binary"
	"encoding/hex"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// invoiceTransferMsg is transferMsg bound to invoiceRef, with a token the fake
// verifier accepts.
func (s *EscrowTestSuite) invoiceTransferMsg(id string, recipient sdk.AccAddress, amt int64, seq uint64, invoiceRef []byte) *types.MsgSubmitAttestedTransfer {
	msg := s.transferMsg(id, recipient, amt, seq)
	msg.InvoiceRef = invoiceRef
	payload := types.AttestedTransferBytes(recipient.String(), spendDenom, msg.Amount, invoiceRef, msg.Memo)
	msg.Token = types.TransferNonce(id, payload, seq, len(invoiceRef) > 0)
	return msg
}

func (s *EscrowTestSuite) invoiceSettlement(agentID string, recipient sdk.AccAddress, invoiceRef []byte) *types.QueryInvoiceSettlementResponse {
	res, err := s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{AgentId: agentID, Recipient: recipient.String(), InvoiceRef: invoiceRef})
	s.Require().NoError(err)
	return res
}

func (s *EscrowTestSuite) TestInvoiceTransfer_SettlesOnce() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	ref := []byte("inv-\x00-7")

	s.Require().False(s.invoiceSettlement("a1", recipient, ref).Settled)

	s.Ctx = s.Ctx.WithBlockHeight(42)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 0, ref))
	s.Require().NoError(err)

	res := s.invoiceSettlement("a1", recipient, ref)
	s.Require().True(res.Settled)
	s.Require().Equal(types.InvoiceSettlement{
		AgentId: "a1",
		Amount:  sdk.NewInt64Coin(spendDenom, 100),
		Height:  42,
		Seq:     0,
	}, res.Settlement)

	events := s.Ctx.EventManager().Events()
	last := events[len(events)-1]
	s.Require().Equal("dymensionxyz.dymension.agent.EventAttestedTransfer", last.Type)
	attr, found := last.GetAttribute("invoice_ref")
	s.Require().True(found)
	s.Require().Equal("aW52LQAtNw==", attr.Value) // base64 of ref

	// a fresh, validly attested transfer at the next seq paying the same
	// invoice again is rejected, leaving escrow, seq and payee untouched
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 1, ref))
	s.Require().ErrorIs(err, types.ErrInvoiceAlreadySettled)

	s.Require().Equal(math.NewInt(400), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().Equal(math.NewInt(100), s.App.BankKeeper.GetBalance(s.Ctx, recipient, spendDenom).Amount)
	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(1), agent.ActionSeq)
	s.Require().Equal(math.NewInt(100), agent.SpendWindowSpent)

	// a different invoice to the same payee still settles
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 1, []byte("inv-8")))
	s.Require().NoError(err)
}

// TestInvoiceTransfer_RecipientScoped pins that, within one agent, settling a
// ref to one payee does not block paying the same ref to another payee.
func (s *EscrowTestSuite) TestInvoiceTransfer_RecipientScoped() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, first := testdata.KeyTestPubAddr()
	_, _, second := testdata.KeyTestPubAddr()
	ref := []byte("shared")

	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", first, 100, 0, ref))
	s.Require().NoError(err)
	s.Require().False(s.invoiceSettlement("a1", second, ref).Settled)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", second, 100, 1, ref))
	s.Require().NoError(err)
	s.Require().True(s.invoiceSettlement("a1", first, ref).Settled)
	s.Require().True(s.invoiceSettlement("a1", second, ref).Settled)
	s.Require().Equal(uint64(1), s.invoiceSettlement("a1", second, ref).Settlement.Seq)
}

// TestInvoiceTransfer_UppercaseRecipientSameNamespace pins that the
// all-uppercase bech32 spelling of a payee shares its invoice namespace.
func (s *EscrowTestSuite) TestInvoiceTransfer_UppercaseRecipientSameNamespace() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	ref := []byte("inv-1")

	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 0, ref))
	s.Require().NoError(err)

	upper := strings.ToUpper(recipient.String())
	msg := s.invoiceTransferMsg("a1", recipient, 100, 1, ref)
	msg.Recipient = upper
	payload := types.AttestedTransferBytes(upper, spendDenom, msg.Amount, ref, msg.Memo)
	msg.Token = types.TransferNonce("a1", payload, 1, true)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrInvoiceAlreadySettled)

	res, err := s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{AgentId: "a1", Recipient: upper, InvoiceRef: ref})
	s.Require().NoError(err)
	s.Require().True(res.Settled)
}

// TestInvoiceTransfer_OtherAgentCannotSquat pins the payer-scoped namespace:
// independently owned agents paying a merchant's invoice ref with the wrong
// amount or the wrong denom do not consume it, so the expected payer still
// settles it exactly once and its receipt reflects only its own payment.
func (s *EscrowTestSuite) TestInvoiceTransfer_OtherAgentCannotSquat() {
	_, _, merchant := testdata.KeyTestPubAddr()
	ref := []byte("inv-123") // issued to a1 for 100adym

	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)

	// b1: wrong amount, same denom
	s.spendingAgent("b1")
	s.fundEscrow("b1", 500)
	_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("b1", merchant, 1, 0, ref))
	s.Require().NoError(err)

	// b2: wrong denom
	const otherDenom = "uatom"
	b2Owner := s.registerAgent("b2")
	_, err = s.msgServer.UpdateAgentSpendPolicy(s.Ctx,
		types.NewMsgUpdateAgentSpendPolicy(b2Owner.String(), "b2", otherDenom, math.NewInt(windowLimit), windowLen))
	s.Require().NoError(err)
	_, _, funder := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewInt64Coin(otherDenom, 500))
	s.FundAcc(funder, coins)
	_, err = s.msgServer.FundAgentEscrow(s.Ctx, types.NewMsgFundAgentEscrow(funder.String(), "b2", coins))
	s.Require().NoError(err)
	b2Msg := s.invoiceTransferMsg("b2", merchant, 1, 0, ref)
	b2Payload := types.AttestedTransferBytes(merchant.String(), otherDenom, b2Msg.Amount, ref, b2Msg.Memo)
	b2Msg.Token = types.TransferNonce("b2", b2Payload, 0, true)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, b2Msg)
	s.Require().NoError(err)

	s.Require().False(s.invoiceSettlement("a1", merchant, ref).Settled)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", merchant, 100, 0, ref))
	s.Require().NoError(err)

	res := s.invoiceSettlement("a1", merchant, ref)
	s.Require().True(res.Settled)
	s.Require().Equal(sdk.NewInt64Coin(spendDenom, 100), res.Settlement.Amount)
	s.Require().Equal(sdk.NewInt64Coin(spendDenom, 1), s.invoiceSettlement("b1", merchant, ref).Settlement.Amount)
	s.Require().Equal(sdk.NewInt64Coin(otherDenom, 1), s.invoiceSettlement("b2", merchant, ref).Settlement.Amount)

	// a1 still cannot pay the same invoice twice
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", merchant, 100, 1, ref))
	s.Require().ErrorIs(err, types.ErrInvoiceAlreadySettled)
}

func (s *EscrowTestSuite) TestTransfer_EmptyInvoiceRefIsLegacy() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	for seq := uint64(0); seq < 2; seq++ {
		_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.transferMsg("a1", recipient, 100, seq))
		s.Require().NoError(err)
	}

	// payload unchanged from the pre-invoice encoding
	entry, found := s.k.GetActionLogEntry(s.Ctx, "a1", 0)
	s.Require().True(found)
	s.Require().Equal([]byte(recipient.String()+"\x00adym\x00100\x00m"), entry.Payload)
	s.Require().Empty(keeper.ExportGenesis(s.Ctx, s.k).InvoiceSettlements)

	events := s.Ctx.EventManager().Events()
	attr, found := events[len(events)-1].GetAttribute("invoice_ref")
	s.Require().True(found)
	s.Require().Equal("null", attr.Value)
}

func (s *EscrowTestSuite) TestInvoiceTransfer_RefNotAttestedRejected() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	ref := []byte("inv-1")

	tests := []struct {
		name   string
		mutate func(*types.MsgSubmitAttestedTransfer)
	}{
		{"different ref", func(m *types.MsgSubmitAttestedTransfer) { m.InvoiceRef = []byte("inv-2") }},
		{"ref attached to legacy token", func(m *types.MsgSubmitAttestedTransfer) {
			*m = *s.transferMsg("a1", recipient, 100, 0)
			m.InvoiceRef = ref
		}},
		// same payload bytes as the attested invoice transfer, resubmitted
		// without the ref to dodge settlement dedup
		{"ref folded into memo", func(m *types.MsgSubmitAttestedTransfer) {
			memo := binary.BigEndian.AppendUint64(nil, uint64(len(ref)))
			memo = append(memo, ref...)
			m.Memo = append(memo, m.Memo...)
			m.InvoiceRef = nil
		}},
	}
	for _, tc := range tests {
		msg := s.invoiceTransferMsg("a1", recipient, 100, 0, ref)
		tc.mutate(msg)
		_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, msg)
		s.Require().ErrorContains(err, "verify attestation", tc.name)
	}

	agent, _ := s.k.GetAgent(s.Ctx, "a1")
	s.Require().Equal(uint64(0), agent.ActionSeq)
	s.Require().Equal(math.NewInt(500), s.k.GetEscrowBalance(s.Ctx, "a1").AmountOf(spendDenom))
	s.Require().False(s.invoiceSettlement("a1", recipient, ref).Settled)
}

func (s *EscrowTestSuite) TestInvoiceTransfer_RefTooLong() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()

	params, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)

	ref := make([]byte, params.InvoiceRefMaxBytes+1)
	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 0, ref))
	s.Require().ErrorContains(err, "invoice ref exceeds max bytes")
	s.Require().Zero(s.verifier.calls)

	_, err = s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 100, 0, ref[1:]))
	s.Require().NoError(err)
}

func (s *EscrowTestSuite) TestInvoiceSettlementQuery_RejectsEmptyKey() {
	_, _, recipient := testdata.KeyTestPubAddr()
	_, err := s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{Recipient: recipient.String(), InvoiceRef: []byte("x")})
	s.Require().ErrorContains(err, "empty agent id")
	_, err = s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{AgentId: "a1", InvoiceRef: []byte("x")})
	s.Require().ErrorContains(err, "recipient")
	_, err = s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{AgentId: "a1", Recipient: "invalid", InvoiceRef: []byte("x")})
	s.Require().ErrorContains(err, "recipient")
	_, err = s.k.InvoiceSettlement(s.Ctx, &types.QueryInvoiceSettlementRequest{AgentId: "a1", Recipient: recipient.String()})
	s.Require().ErrorContains(err, "empty invoice ref")
}

func (s *EscrowTestSuite) TestInvoiceSettlement_GenesisRoundTrip() {
	s.spendingAgent("a1")
	s.fundEscrow("a1", 500)
	_, _, recipient := testdata.KeyTestPubAddr()
	refs := [][]byte{[]byte("b"), {0x00, 0xff}, []byte("a")}
	for seq, ref := range refs {
		_, err := s.msgServer.SubmitAttestedTransfer(s.Ctx, s.invoiceTransferMsg("a1", recipient, 10, uint64(seq), ref))
		s.Require().NoError(err)
	}

	exported := keeper.ExportGenesis(s.Ctx, s.k)
	s.Require().NoError(exported.Validate())
	s.Require().Len(exported.InvoiceSettlements, len(refs))

	// re-import into a fresh store and export again: identical, and live
	s.SetupTest()
	keeper.InitGenesis(s.Ctx, s.k, *exported)
	s.Require().Equal(exported.InvoiceSettlements, keeper.ExportGenesis(s.Ctx, s.k).InvoiceSettlements)
	for _, ref := range refs {
		res := s.invoiceSettlement("a1", recipient, ref)
		s.Require().True(res.Settled, hex.EncodeToString(ref))
		s.Require().Equal("a1", res.Settlement.AgentId)
	}
}

func (s *EscrowTestSuite) TestMigrate3to4_DefaultsInvoiceRefMaxBytes() {
	params := types.DefaultParams()
	params.InvoiceRefMaxBytes = 0
	s.Require().NoError(s.k.SetParams(s.Ctx, params))

	migrator := keeper.NewMigrator(*s.k)
	s.Require().NoError(migrator.Migrate3to4(s.Ctx))

	params, err := s.k.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(128), params.InvoiceRefMaxBytes)
}
