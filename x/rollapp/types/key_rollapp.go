package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// RollappKeyPrefix is the prefix to retrieve all Rollapp
	RollappKeyPrefix             = "Rollapp/value/"
	RollappByEIP155KeyPrefix     = "RollappByEIP155/value/"
	RollappByIBCChannelKeyPrefix = "RollappByIBCChannel/value/"
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

// RollappByEIP155Key returns the store key to retrieve a Rollapp from the index fields
func RollappByEIP155Key(
	eip155 uint64,
) []byte {
	var key []byte

	eip155Bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(eip155Bytes, eip155)
	key = append(key, eip155Bytes...)
	key = append(key, []byte("/")...)

	return key
}

// RollappByIBCChannelKey returns the store key to retrieve a Rollapp from the index fields
func RollappByIBCChannelKey(
	portID string,
	channelID string,
) []byte {
	var key []byte

	portIDBytes := []byte(portID)
	key = append(key, portIDBytes...)
	key = append(key, []byte("/")...)
	channelIDBytes := []byte(channelID)
	key = append(key, channelIDBytes...)
	key = append(key, []byte("/")...)

	return key
}
