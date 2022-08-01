package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// SequencersByRollappKeyPrefix is the prefix to retrieve all SequencersByRollapp
	SequencersByRollappKeyPrefix = "SequencersByRollapp/value/"
)

// SequencersByRollappKey returns the store key to retrieve a SequencersByRollapp from the index fields
func SequencersByRollappKey(
	rollappId string,
) []byte {
	var key []byte

	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
