package types

import (
	"encoding/binary"
	"fmt"
)

var _ binary.ByteOrder

const (
	// TimestampedStateInfoKeyPrefix is the prefix to retrieve all StateInfo with the timestamp prefix
	TimestampedStateInfoKeyPrefix = "TimestampedStateInfoKeyPrefix/value/"
	// TimestampPrefixLen is the length of the timestamp prefix: len(fmt.Sprint(time.Time{}.UnixMicro())) + 1
	TimestampPrefixLen = 17
)

// StateInfoTimestampKeyPrefix returns the store key prefix to range over all state infos with the timestamp prefix
func StateInfoTimestampKeyPrefix(timestampUNIX int64) []byte {
	return []byte(fmt.Sprint(timestampUNIX))
}

// StateInfoTimestampKey returns the store key to retrieve state infos using the timestamp prefix
func StateInfoTimestampKey(
	stateInfo StateInfo,
) []byte {
	var key []byte

	timestampPrefix := StateInfoTimestampKeyPrefix(stateInfo.CreatedAt.UnixMicro())
	stateInfoKey := StateInfoKey(stateInfo.StateInfoIndex)
	key = append(key, timestampPrefix...)
	key = append(key, []byte("/")...)
	key = append(key, stateInfoKey...)

	return key
}
