package types

import (
	"encoding/binary"
	fmt "fmt"
)

var _ binary.ByteOrder

const (
	// RollappStateInfoKeyPrefix is the prefix to retrieve all RollappStateInfo
	RollappStateInfoKeyPrefix = "RollappStateInfo/value/"
)

// RollappStateInfoKey returns the store key to retrieve a RollappStateInfo from the index fields
func RollappStateInfoKey(
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
