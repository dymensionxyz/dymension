package types

import (
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewForwardMemo(t *testing.T) {
	eibcFee := "100"
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	destinationDomain := uint32(1)
	recipient, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	amount := math.NewInt(100)
	maxFee := sdk.NewCoin("adym", math.NewInt(100))
	recoveryAddr := "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"
	gasLimit := math.ZeroInt()
	var customHookId *hyperutil.HexAddress
	customHookMetadata := ""

	_, err := NewForwardMemo(
		eibcFee,
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		recoveryAddr,
		gasLimit,
		customHookId,
		customHookMetadata,
	)
	require.NoError(t, err)

}
