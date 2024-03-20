package sample

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccAddress returns a sample account address
func AccAddress() string {
	return Acc().String()
}

func Acc() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

// GenerateAddresses generates numOfAddresses bech32 address
func GenerateAddresses(numOfAddresses int) []string {
	addresses := []string{}
	for i := 0; i < numOfAddresses; i++ {
		addresses = append(addresses, AccAddress())
	}
	return addresses
}
