package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	RollappClientKey       = []byte{0x01}
	canonicalClientKey     = []byte{0x04}
	_                      = []byte{0x05}
	HeaderSignersPrefixKey = collections.NewPrefix("headerSigners/")
	ClientHeightToSigner   = collections.NewPrefix("clientHeightToSigner/")
)

func GetRollappClientKey(rollappId string) []byte {
	key := RollappClientKey
	key = append(key, []byte(rollappId)...)
	return key
}

func CanonicalClientKey(clientID string) []byte {
	key := canonicalClientKey
	key = append(key, []byte(clientID)...)
	return key
}
