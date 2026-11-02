package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/crypto/hd"
	"github.com/stretchr/testify/require"
)

func TestKeyringOption(t *testing.T) {
	var options keyring.Options
	keyringOption()(&options)

	require.Equal(t, hd.SupportedAlgorithms, options.SupportedAlgos)
	require.Equal(t, keyring.SigningAlgoList{hd.EthSecp256k1}, options.SupportedAlgosLedger)
	require.NotNil(t, options.LedgerDerivation)
	require.Equal(t, "Cosmos", options.LedgerAppName)
	require.False(t, options.LedgerSigSkipDERConv, "the Cosmos app returns DER signatures")

	pubKey := options.LedgerCreateKey(make([]byte, ethsecp256k1.PubKeySize))
	require.IsType(t, &ethsecp256k1.PubKey{}, pubKey)
}
