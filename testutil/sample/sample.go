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

func AccAddressFromSecret(secret string) string {
	return AccFromSecret(secret).String()
}

func AccFromSecret(secret string) sdk.AccAddress {
	pk := ed25519.GenPrivKeyFromSecret([]byte(secret)).PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}
