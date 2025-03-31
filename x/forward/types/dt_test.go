package types

import (
	"testing"

	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
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

func TestNewHyperlaneMessage(t *testing.T) {

	srcContract, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")

	_, err := NewHyperlaneMessage(
		1,
		1,
		srcContract,
		0,
		tokenId,
		sample.Acc(),
		math.NewInt(100),
		"channel-0",
		"ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk",
		sdk.NewCoin("adym", math.NewInt(100)),
		1000000000000000000,
		sample.AccAddress(),
	)
	require.NoError(t, err)
}
