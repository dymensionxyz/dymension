package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// SequencerKeyPrefix is the prefix to retrieve all Sequencer
	SequencerKeyPrefix = "Sequencer/value/"
)

// SequencerKey returns the store key to retrieve a Sequencer from the index fields
func SequencerKey(
	sequencerAddress string,
) []byte {
	var key []byte

	sequencerAddressBytes := []byte(sequencerAddress)
	key = append(key, sequencerAddressBytes...)
	key = append(key, []byte("/")...)

	return key
}
