package sample

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccAddress returns a sample account address
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr).String()
}

// GenerateAddresses generates numOfAddresses bech32 address
func GenerateAddresses(numOfAddresses int) []string {
	addresses := []string{}
	for i := 0; i < numOfAddresses; i++ {
		addresses = append(addresses, AccAddress())
	}
	return addresses
}
