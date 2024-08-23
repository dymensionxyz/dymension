package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

// KV Store
var (
	rollappClientKey        = []byte{0x01}
	consensusStateSignerKey = []byte{0x03}
	canonicalClientKey      = []byte{0x04}
)

func RollappClientKey(rollappId string) []byte {
	return append(rollappClientKey, []byte(rollappId)...)
}

func ConsensusStateSignerKeyByClientID(clientID string, height uint64) []byte {
	prefix := append([]byte(clientID), sdk.Uint64ToBigEndian(height)...)
	return append(consensusStateSignerKey, prefix...)
}

func CanonicalClientKey(clientID string) []byte {
	return append(canonicalClientKey, []byte(clientID)...)
}
