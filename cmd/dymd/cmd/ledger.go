package cmd

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/crypto/hd"

	"github.com/dymensionxyz/dymension/v3/internal/ledgercosmos"
)

// keyringOption supports eth_secp256k1 and secp256k1 keys, and makes --ledger keys use the
// Ledger Cosmos app with the Dymension chain config: eth_secp256k1 keys whose address is the
// one shown on the device (and in Keplr and MetaMask). See package ledgercosmos.
func keyringOption() keyring.Option {
	derive := ledgercosmos.Discover(func() string { return sdk.GetConfig().GetBech32AccountAddrPrefix() })

	return func(options *keyring.Options) {
		hd.EthSecp256k1Option()(options)
		options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.EthSecp256k1}
		options.LedgerDerivation = derive
		options.LedgerCreateKey = func(key []byte) cryptotypes.PubKey { return &ethsecp256k1.PubKey{Key: key} }
		options.LedgerAppName = "Cosmos"
	}
}
