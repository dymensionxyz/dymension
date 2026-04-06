package kaspaaddr

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/stretchr/testify/require"
)

func TestFromH256(t *testing.T) {
	// Test case: H256 -> Kaspa address
	// Verified with Rust kaspa-tools: cargo run -p kaspa-tools -- recipient kaspatest:qzlq49spp66vkjjex0w7z8708f6zteqwr6swy33fmy4za866ne90vhy54uh3j
	h256Hex := "be0a96010eb4cb4a5933dde11fcf3a7425e40e1ea0e24629d92a2e9f5a9e4af6"
	expectedAddr := "kaspatest:qzlq49spp66vkjjex0w7z8708f6zteqwr6swy33fmy4za866ne90vhy54uh3j"

	h256, err := hex.DecodeString(h256Hex)
	require.NoError(t, err)
	require.Len(t, h256, 32)

	addr, err := FromH256(h256, false) // testnet
	require.NoError(t, err)
	require.Equal(t, expectedAddr, addr)
}

func TestFromH256Mainnet(t *testing.T) {
	h256Hex := "be0a96010eb4cb4a5933dde11fcf3a7425e40e1ea0e24629d92a2e9f5a9e4af6"

	h256, err := hex.DecodeString(h256Hex)
	require.NoError(t, err)

	addr, err := FromH256(h256, true) // mainnet
	require.NoError(t, err)
	require.True(t, len(addr) > 0)
	require.Contains(t, addr, "kaspa:")
	require.NotContains(t, addr, "kaspatest:")
}

func TestConvertBitsRoundtrip(t *testing.T) {
	input := []byte{0x00, 0xbe, 0x0a}

	result5, err := bech32.ConvertBits(input, 8, 5, true)
	require.NoError(t, err)
	require.NotEmpty(t, result5)

	result8, err := bech32.ConvertBits(result5, 5, 8, false)
	require.NoError(t, err)
	require.Equal(t, input, result8)
}

func TestFromH256InvalidLength(t *testing.T) {
	_, err := FromH256([]byte{0x01, 0x02, 0x03}, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid H256 length")

	_, err = FromH256(make([]byte, 33), true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid H256 length")
}
