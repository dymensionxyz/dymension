package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LatestFinalizedStateIndexKeyPrefix is the prefix to retrieve all LatestFinalizedStateIndex
	LatestFinalizedStateIndexKeyPrefix = "LatestFinalizedStateIndex/value/"
)

// LatestFinalizedStateIndexKey returns the store key to retrieve a LatestFinalizedStateIndex from the index fields
func LatestFinalizedStateIndexKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
