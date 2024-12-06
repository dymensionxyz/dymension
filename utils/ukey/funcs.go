package ukey

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RandomTMPubKey() cryptotypes.PubKey {
	return ed25519.GenPrivKey().PubKey()
}

func PkAcc(pk cryptotypes.PubKey) sdk.AccAddress {
	return sdk.AccAddress(pk.Address())
}

func PkAddr(pk cryptotypes.PubKey) string {
	return PkAcc(pk).String()
}
