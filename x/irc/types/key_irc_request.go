package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// IRCRequestKeyPrefix is the prefix to retrieve all IRCRequest
	IRCRequestKeyPrefix = "IRCRequest/value/"
)

// IRCRequestKey returns the store key to retrieve a IRCRequest from the index fields
func IRCRequestKey(
	reqId uint64,
) []byte {
	var key []byte

	reqIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reqIdBytes, reqId)
	key = append(key, reqIdBytes...)
	key = append(key, []byte("/")...)

	return key
}
