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
