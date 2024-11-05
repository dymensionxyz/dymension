package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// BlockHeightToFinalizationQueueKeyPrefix is the prefix to retrieve all BlockHeightToFinalizationQueue
	BlockHeightToFinalizationQueueKeyPrefix = "BlockHeightToFinalizationQueue/value/"

	// HeightRollappToFinalizationQueueKeyPrefix is the prefix to retrieve all FinalizationQueue by (height, rollappID)
	HeightRollappToFinalizationQueueKeyPrefix = "HeightRollappToFinalizationQueue/value/"
	// RollappHeightToFinalizationQueueKeyPrefix is the prefix to retrieve all FinalizationQueue by (rollappID, height)
	RollappHeightToFinalizationQueueKeyPrefix = "RollappHeightToFinalizationQueue/value/"
)

// BlockHeightToFinalizationQueueKey returns the store key to retrieve a BlockHeightToFinalizationQueue from the index fields
func BlockHeightToFinalizationQueueKey(
	creationHeight uint64,
) []byte {
	var key []byte

	creationHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(creationHeightBytes, creationHeight)
	key = append(key, creationHeightBytes...)
	key = append(key, []byte("/")...)

	return key
}
