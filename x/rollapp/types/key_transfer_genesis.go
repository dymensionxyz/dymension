package types

import (
	"encoding/binary"
)

var _ binary.ByteOrder

const (
	TransferGenesisMapKeyPrefix = "TransferGenesis/map/value/"
)

var (
	transferGenesisNumTotalSubkey = []byte{1}
	transferGenesisNumSubkey      = []byte{2}
)

// TransferGenesisNumTotalKey returns the store key to check the total number of genesis transfers that the rollapp has decided to do
func TransferGenesisNumTotalKey(
	rollappID string,
) []byte {
	var key []byte
	// build the key bytes
	rollappIdBytes := []byte(rollappID)
	// concatenate the byte slices directly
	key = append(key, transferGenesisNumTotalSubkey...)
	key = append(key, []byte("/")...)
	key = append(key, rollappIdBytes...)

	return key
}

// TransferGenesisNumKey returns the store key to check the number of genesis transfers that the rollapp has done
func TransferGenesisNumKey(
	rollappID string,
) []byte {
	var key []byte
	// build the key bytes
	rollappIdBytes := []byte(rollappID)
	// concatenate the byte slices directly
	key = append(key, transferGenesisNumSubkey...)
	key = append(key, []byte("/")...)
	key = append(key, rollappIdBytes...)

	return key
}
