package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// RollappKeyPrefix is the prefix to retrieve all Rollapp
	RollappKeyPrefix = "Rollapp/value/"
)

// RollappKey returns the store key to retrieve a Rollapp from the index fields
func RollappKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
