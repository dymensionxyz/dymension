package cli

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestH256ToKaspaAddress(t *testing.T) {
	// Test case: H256 -> Kaspa address roundtrip
	// This H256 corresponds to: kaspatest:qzlq49spp66vkjjex0w7z8708f6zteqwr6swy33fmy4za866ne90vhy54uh3j
	// Verified with Rust kaspa-tools: cargo run -p kaspa-tools -- recipient kaspatest:qzlq49spp66vkjjex0w7z8708f6zteqwr6swy33fmy4za866ne90vhy54uh3j
	h256Hex := "be0a96010eb4cb4a5933dde11fcf3a7425e40e1ea0e24629d92a2e9f5a9e4af6"
	expectedAddr := "kaspatest:qzlq49spp66vkjjex0w7z8708f6zteqwr6swy33fmy4za866ne90vhy54uh3j"

	h256, err := hex.DecodeString(h256Hex)
	require.NoError(t, err)
	require.Len(t, h256, 32)

	addr, err := H256ToKaspaAddress(h256, false) // testnet
	require.NoError(t, err)
	require.Equal(t, expectedAddr, addr)
}

func TestH256ToKaspaAddressMainnet(t *testing.T) {
	// Same payload but mainnet prefix
	h256Hex := "be0a96010eb4cb4a5933dde11fcf3a7425e40e1ea0e24629d92a2e9f5a9e4af6"

	h256, err := hex.DecodeString(h256Hex)
	require.NoError(t, err)

	addr, err := H256ToKaspaAddress(h256, true) // mainnet
	require.NoError(t, err)
	require.True(t, len(addr) > 0)
	require.Contains(t, addr, "kaspa:")
	require.NotContains(t, addr, "kaspatest:")
}

func TestConvertBits(t *testing.T) {
	// Simple test: convert some bytes from 8-bit to 5-bit and back
	input := []byte{0x00, 0xbe, 0x0a}

	// 8 -> 5
	result5, err := convertBits(input, 8, 5, true)
	require.NoError(t, err)
	require.NotEmpty(t, result5)

	// 5 -> 8 (round trip won't be exact due to padding)
	result8, err := convertBits(result5, 5, 8, false)
	require.NoError(t, err)

	// First bytes should match
	require.Equal(t, input, result8)
}

func TestH256ToKaspaAddressInvalidLength(t *testing.T) {
	// Too short
	_, err := H256ToKaspaAddress([]byte{0x01, 0x02, 0x03}, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid H256 length")

	// Too long
	_, err = H256ToKaspaAddress(make([]byte, 33), true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid H256 length")
}
