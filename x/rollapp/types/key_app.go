package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// AppKeyPrefix is the prefix to retrieve all App
	AppKeyPrefix = "App/value/"
)

// AppKey returns the store key to retrieve an App from the index fields
func AppKey(name, rollappId string) []byte {
	var key []byte

	appNameBytes := []byte(name)
	key = append(key, appNameBytes...)
	key = append(key, []byte("/")...)
	rollappIdBytes := []byte(rollappId)
	key = append(key, rollappIdBytes...)

	return key
}
