package types_test

import (
	"bytes"
	"testing"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/stretchr/testify/require"
)

func TestValidationNonceCommitsFullVerdict(t *testing.T) {
	hash := bytes.Repeat([]byte{1}, 32)
	evidence := bytes.Repeat([]byte{2}, 32)
	payload := types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass", "subject", 1)
	nonce := types.ValidationNonce("validator", payload, 7)
	require.Equal(t, nonce, types.ValidationNonce("validator", types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass", "subject", 1), 7))
	variants := [][]byte{
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass", "other-subject", 1),
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass", "subject", 2),
		types.AttestedValidationBytes(bytes.Repeat([]byte{3}, 32), 100, evidence, "ipfs://evidence", "pass", "subject", 1),
		types.AttestedValidationBytes(hash, 99, evidence, "ipfs://evidence", "pass", "subject", 1),
		types.AttestedValidationBytes(hash, 100, nil, "ipfs://evidence", "pass", "subject", 1),
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://other", "pass", "subject", 1),
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "fail", "subject", 1),
	}
	for _, other := range variants {
		require.NotEqual(t, nonce, types.ValidationNonce("validator", other, 7))
	}
	require.NotEqual(t, nonce, types.ValidationNonce("other", payload, 7))
	require.NotEqual(t, nonce, types.ValidationNonce("validator", payload, 8))
	require.NotEqual(t, nonce, types.ActionNonce("validator", payload, 7))
	require.NotEqual(t, nonce, types.TransferNonce("validator", payload, 7))
	require.NotEqual(t, types.AttestedValidationBytes(hash, 0, nil, "a\x00b", "c", "subject", 1), types.AttestedValidationBytes(hash, 0, nil, "a", "b\x00c", "subject", 1))
	require.NotEqual(t, types.AttestedValidationBytes(hash, 0, nil, "", "", "subject", 1), types.AttestedValidationBytes(hash, 0, make([]byte, 32), "", "", "subject", 1))
}

func TestTransferNonceCommitsDeadline(t *testing.T) {
	amount := math.NewInt(42)
	payload := types.AttestedTransferBytes("dym1r", "adym", amount, 0x0102030405060708, []byte("memo"))
	require.Equal(t, []byte("dym1r\x00adym\x0042\x00\x01\x02\x03\x04\x05\x06\x07\x08memo"), payload)

	nonce := types.TransferNonce("agent", payload, 7)
	require.Equal(t, nonce, types.TransferNonce("agent", types.AttestedTransferBytes("dym1r", "adym", amount, 0x0102030405060708, []byte("memo")), 7))
	require.NotEqual(t, nonce, types.TransferNonce("agent", types.AttestedTransferBytes("dym1r", "adym", amount, 0x0102030405060709, []byte("memo")), 7))
}
