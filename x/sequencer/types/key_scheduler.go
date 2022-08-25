package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // SchedulerKeyPrefix is the prefix to retrieve all Scheduler
	SchedulerKeyPrefix = "Scheduler/value/"
)

// SchedulerKey returns the store key to retrieve a Scheduler from the index fields
func SchedulerKey(
sequencerAddress string,
) []byte {
	var key []byte
    
    sequencerAddressBytes := []byte(sequencerAddress)
    key = append(key, sequencerAddressBytes...)
    key = append(key, []byte("/")...)
    
	return key
}