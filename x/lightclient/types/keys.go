package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// TransientKey defines the module's transient store key
	TransientKey = "t_lightclient"
)

// KV Store
var (
	rollappClientKey        = []byte{0x01}
	consensusStateSignerKey = []byte{0x03}
	canonicalClientKey      = []byte{0x04}
)

// Transient Store
var (
	lightClientRegistrationKey = []byte{0x02}
)

func RollappClientKey(rollappId string) []byte {
	return append(rollappClientKey, []byte(rollappId)...)
}

func CanonicalLightClientRegistrationKey(rollappId string) []byte {
	return append(lightClientRegistrationKey, []byte(rollappId)...)
}

func ConsensusStateSignerKeyByClientID(clientID string, height uint64) []byte {
	prefix := append([]byte(clientID), sdk.Uint64ToBigEndian(height)...)
	return append(consensusStateSignerKey, prefix...)
}

func CanonicalClientKey(clientID string) []byte {
	return append(canonicalClientKey, []byte(clientID)...)
}
