package types

import (
	"encoding/binary"
)

var _ binary.ByteOrder

const (
	// AppKeyPrefix is the prefix to retrieve all App
	AppKeyPrefix         = "App/value/"
	AppSequenceKeyPrefix = "App/sequence/"
)

// AppKey returns the store key to retrieve an App from the index fields
func AppKey(app App) []byte {
	var key []byte

	rollappIDBytes := []byte(app.RollappId)
	key = append(key, rollappIDBytes...)
	key = append(key, []byte("/")...)
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, app.Id)
	key = append(key, idBytes...)

	return key
}

func RollappAppKeyPrefix(rollappId string) []byte {
	return append([]byte(rollappId), []byte("/")...)
}

func AppSequenceKey(rollappId string) []byte {
	return []byte(rollappId)
}
