package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LatestFinalizedStateIndexKeyPrefix is the prefix to retrieve all LatestFinalizedStateIndex
	LatestFinalizedStateGlobalIndexKeyPrefix = "LatestFinalizedGlobalStateIndex/value/"
)

// LatestFinalizedStateIndexKey returns the store key to retrieve a LatestFinalizedStateIndex from the index fields
func LatestFinalizedStateGlobalIndexKey() []byte {
	var key []byte

	key = append(key, []byte("/")...)

	return key
}
