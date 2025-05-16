package cli

import (
	"math/big"
	"testing"

	math "cosmossdk.io/math"
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/stretchr/testify/require"
)

func TestMakeForwardToIBCHyperlaneMessage(t *testing.T) {

	srcContract, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")
	tokenId, _ := hyperutil.DecodeHexAddress("0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0")

	_, err := MakeForwardToIBCHyperlaneMessage(
		1,
		1,
		srcContract,
		0,
		tokenId,
		sample.Acc(),
		math.NewInt(100),
		forwardtypes.NewHookForwardToIBC(
			"channel-0",
			"ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk",
			1000000000000000000,
		),
	)
	require.NoError(t, err)
}

// does roundtrip to check we get what we expect in the end
func TestEthRecipient(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(privKey.PubKey().Address())
	addrS := addr.String()

	ans, err := EthRecipient(addrS)
	require.NoError(t, err)

	// mimic decodings
	decoded, err := util.DecodeEthHex(ans)
	require.NoError(t, err)
	pl, err := types.NewWarpPayload(decoded, big.Int{}, nil)
	require.NoError(t, err)
	addrSAfter := pl.GetCosmosAccount().String()

	require.Equal(t, addrSAfter, addrS)
}
