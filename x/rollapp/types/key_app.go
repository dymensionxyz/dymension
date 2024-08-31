package types

import (
	"encoding/binary"
)

var _ binary.ByteOrder

const (
	// AppKeyPrefix is the prefix to retrieve all App
	AppKeyPrefix = "App/value/"
)

// AppKey returns the store key to retrieve an App from the index fields
func AppKey(app App) []byte {
	var key []byte

	rollappIDBytes := []byte(app.RollappId)
	key = append(key, rollappIDBytes...)
	key = append(key, []byte("/")...)
	appNameBytes := []byte(app.Name)
	key = append(key, appNameBytes...)

	return key
}

func RollappAppKeyPrefix(rollappId string) []byte {
	return append([]byte(rollappId), []byte("/")...)
}
