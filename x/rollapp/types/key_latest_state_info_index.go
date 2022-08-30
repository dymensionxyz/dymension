package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LatestStateInfoIndexKeyPrefix is the prefix to retrieve all LatestStateInfoIndex
	LatestStateInfoIndexKeyPrefix = "LatestStateInfoIndex/value/"
)

// LatestStateInfoIndexKey returns the store key to retrieve a LatestStateInfoIndex from the index fields
func LatestStateInfoIndexKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
