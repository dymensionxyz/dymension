package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // BlockHeightToFinalizationQueueKeyPrefix is the prefix to retrieve all BlockHeightToFinalizationQueue
	BlockHeightToFinalizationQueueKeyPrefix = "BlockHeightToFinalizationQueue/value/"
)

// BlockHeightToFinalizationQueueKey returns the store key to retrieve a BlockHeightToFinalizationQueue from the index fields
func BlockHeightToFinalizationQueueKey(
finalizationHeight uint64,
) []byte {
	var key []byte
    
    finalizationHeightBytes := make([]byte, 8)
  					binary.BigEndian.PutUint64(finalizationHeightBytes, finalizationHeight)
    key = append(key, finalizationHeightBytes...)
    key = append(key, []byte("/")...)
    
	return key
}