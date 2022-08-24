package types

import (
	"encoding/binary"
	fmt "fmt"
)

var _ binary.ByteOrder

const (
	// StateInfoKeyPrefix is the prefix to retrieve all StateInfo
	StateInfoKeyPrefix = "StateInfo/value/"
)

// StateInfoKey returns the store key to retrieve a StateInfo from the index fields
func StateInfoKey(
	rollappId string,
	stateIndex uint64,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)
	stateIndexBytes := []byte(fmt.Sprint(stateIndex))
	key = append(key, stateIndexBytes...)
	key = append(key, []byte("/")...)

	return key
}
