package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	// StateInfoKeyPrefix is the prefix to retrieve all StateInfo
	StateInfoKeyPrefix = "StateInfo/value/"
)

// StateInfoKey returns the store key to retrieve a StateInfo from the index fields
func StateInfoKey(
	stateInfoIndex StateInfoIndex,
) []byte {
	var key []byte
	// build the key bytes
	rollappIdBytes := []byte(stateInfoIndex.RollappId)
	stateInfoIndexBytes := sdk.Uint64ToBigEndian(stateInfoIndex.Index)
	// concatenate the byte slices directly
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)
	key = append(key, stateInfoIndexBytes...)
	key = append(key, []byte("/")...)

	return key
}
