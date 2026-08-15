package cli

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

func TestUpdateAgentSpendPolicyCommand(t *testing.T) {
	_, _, first := testdata.KeyTestPubAddr()
	_, _, second := testdata.KeyTestPubAddr()

	t.Run("requires explicit recipient allowlist", func(t *testing.T) {
		cmd := CmdUpdateAgentSpendPolicy()
		require.NoError(t, cmd.Args(cmd, []string{"agent-1", "adym", "100", "10", first.String()}))
		require.Error(t, cmd.Args(cmd, []string{"agent-1", "adym", "100", "10"}))
	})

	t.Run("constructs recipient allowlist", func(t *testing.T) {
		msg, err := newMsgUpdateAgentSpendPolicy(
			"dym1owner",
			[]string{"agent-1", "adym", "100", "10", first.String() + "," + second.String()},
		)
		require.NoError(t, err)
		require.Equal(t, []string{first.String(), second.String()}, msg.RecipientAllowlist)
	})

	t.Run("empty recipient allowlist is explicit unrestricted policy", func(t *testing.T) {
		msg, err := newMsgUpdateAgentSpendPolicy("dym1owner", []string{"agent-1", "adym", "100", "10", ""})
		require.NoError(t, err)
		require.Empty(t, msg.RecipientAllowlist)
	})
}

func TestSubmitAttestedTransferCommand(t *testing.T) {
	t.Run("registered with exact arguments", func(t *testing.T) {
		cmd, _, err := GetTxCmd().Find([]string{"submit-attested-transfer"})
		require.NoError(t, err)
		require.Equal(t, "submit-attested-transfer", cmd.Name())
		require.NoError(t, cmd.Args(cmd, []string{"agent-1", "dym1recipient", "42", "invoice-7", "token"}))
		require.Error(t, cmd.Args(cmd, []string{"agent-1", "dym1recipient", "42", "invoice-7"}))
		require.Error(t, cmd.Args(cmd, []string{"agent-1", "dym1recipient", "42", "invoice-7", "token", "extra"}))
	})

	t.Run("constructs message from arguments", func(t *testing.T) {
		msg, err := newMsgSubmitAttestedTransfer(
			"dym1submitter",
			[]string{"agent-1", "dym1recipient", "42", "invoice-7", "attestation-token"},
		)
		require.NoError(t, err)
		require.Equal(t, "dym1submitter", msg.Submitter)
		require.Equal(t, "agent-1", msg.AgentId)
		require.Equal(t, "dym1recipient", msg.Recipient)
		require.True(t, math.NewInt(42).Equal(msg.Amount))
		require.Equal(t, []byte("invoice-7"), msg.Memo)
		require.Equal(t, "attestation-token", msg.Token)
	})

	t.Run("rejects invalid amount", func(t *testing.T) {
		_, err := newMsgSubmitAttestedTransfer(
			"dym1submitter",
			[]string{"agent-1", "dym1recipient", "not-an-integer", "memo", "token"},
		)
		require.ErrorContains(t, err, "amount")
	})
}
