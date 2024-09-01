package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

const (
	keySeparator = "/"
)

var (
	RollappClientKey         = []byte{0x01}
	ConsensusStateValhashKey = []byte{0x03}
	canonicalClientKey       = []byte{0x04}
)

func GetRollappClientKey(rollappId string) []byte {
	key := RollappClientKey
	key = append(key, []byte(rollappId)...)
	return key
}

func ConsensusStateValhashKeyByClientID(clientID string, height uint64) []byte {
	key := ConsensusStateValhashKey
	key = append(key, []byte(clientID)...)
	key = append(key, keySeparator...)
	key = append(key, sdk.Uint64ToBigEndian(height)...)
	return key
}

func CanonicalClientKey(clientID string) []byte {
	key := canonicalClientKey
	key = append(key, []byte(clientID)...)
	return key
}

func ParseConsensusStateValhashKey(key []byte) (clientID string, height uint64) {
	key = key[len(ConsensusStateValhashKey):]
	parts := bytes.Split(key, []byte(keySeparator))
	clientID = string(parts[0])
	height = sdk.BigEndianToUint64(parts[1])
	return
}
