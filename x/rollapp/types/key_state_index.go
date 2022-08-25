package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// StateIndexKeyPrefix is the prefix to retrieve all StateIndex
	StateIndexKeyPrefix = "StateIndex/value/"
)

// StateIndexKey returns the store key to retrieve a StateIndex from the index fields
func StateIndexKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
