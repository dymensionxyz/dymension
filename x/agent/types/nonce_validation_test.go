package types_test

import (
	"bytes"
	"testing"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/stretchr/testify/require"
)

func TestValidationNonceCommitsFullVerdict(t *testing.T) {
	hash := bytes.Repeat([]byte{1}, 32)
	evidence := bytes.Repeat([]byte{2}, 32)
	payload := types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass")
	nonce := types.ValidationNonce("validator", payload, 7)
	require.Equal(t, nonce, types.ValidationNonce("validator", types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "pass"), 7))
	variants := [][]byte{
		types.AttestedValidationBytes(bytes.Repeat([]byte{3}, 32), 100, evidence, "ipfs://evidence", "pass"),
		types.AttestedValidationBytes(hash, 99, evidence, "ipfs://evidence", "pass"),
		types.AttestedValidationBytes(hash, 100, nil, "ipfs://evidence", "pass"),
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://other", "pass"),
		types.AttestedValidationBytes(hash, 100, evidence, "ipfs://evidence", "fail"),
	}
	for _, other := range variants {
		require.NotEqual(t, nonce, types.ValidationNonce("validator", other, 7))
	}
	require.NotEqual(t, nonce, types.ValidationNonce("other", payload, 7))
	require.NotEqual(t, nonce, types.ValidationNonce("validator", payload, 8))
	require.NotEqual(t, nonce, types.ActionNonce("validator", payload, 7))
	require.NotEqual(t, nonce, types.TransferNonce("validator", payload, 7))
	require.NotEqual(t, types.AttestedValidationBytes(hash, 0, nil, "a\x00b", "c"), types.AttestedValidationBytes(hash, 0, nil, "a", "b\x00c"))
	require.NotEqual(t, types.AttestedValidationBytes(hash, 0, nil, "", ""), types.AttestedValidationBytes(hash, 0, make([]byte, 32), "", ""))
}
