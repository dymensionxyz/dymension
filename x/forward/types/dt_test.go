package types

import (
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMakeRolForwardToHLMemoString(t *testing.T) {
	eibcFee := "100"
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	destinationDomain := uint32(1)
	recipient, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	amount := math.NewInt(100)
	maxFee := sdk.NewCoin("adym", math.NewInt(100))
	gasLimit := math.ZeroInt()
	var customHookId *hyperutil.HexAddress
	customHookMetadata := ""

	hook := NewHookForwardToHL(
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		gasLimit,
		customHookId,
		customHookMetadata,
	)

	_, err := MakeRolForwardToHLMemoString(eibcFee, hook)
	require.NoError(t, err)
}

func TestHLMetadataWithMultipleForwards(t *testing.T) {
	// Test that HLMetadata can contain multiple forward types without conflict
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")

	// Create HL-to-HL forward data
	hlForward := NewHookForwardToHL(
		tokenId,
		uint32(42161), // Arbitrum
		tokenId,
		math.NewInt(1000),
		sdk.NewCoin("adym", math.NewInt(10)),
		math.NewInt(100000),
		nil,
		"",
	)
	hlForwardBz, err := hlForward.Marshal()
	require.NoError(t, err)

	// Create HL-to-IBC forward data
	ibcForward := NewHookForwardToIBC(
		"channel-0",
		"cosmos1...",
		1234567890,
	)
	ibcForwardBz, err := ibcForward.Marshal()
	require.NoError(t, err)

	// Create metadata with both forwards (though in practice only one would be used)
	metadata := &HLMetadata{
		HookForwardToIbc: ibcForwardBz,
		HookForwardToHl:  hlForwardBz,
	}

	// Marshal and unmarshal
	metadataBz, err := metadata.Marshal()
	require.NoError(t, err)

	unpackedMetadata, err := UnpackHLMetadata(metadataBz)
	require.NoError(t, err)
	require.NotNil(t, unpackedMetadata)
	require.Equal(t, ibcForwardBz, unpackedMetadata.HookForwardToIbc)
	require.Equal(t, hlForwardBz, unpackedMetadata.HookForwardToHl)
}
