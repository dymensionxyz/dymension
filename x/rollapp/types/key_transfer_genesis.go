package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	TransferGenesisMapKeyPrefix = "TransferGenesis/map/value/"
)

var (
	transferGenesisSetMembershipSubkey = []byte{0}
	transferGenesisNumTotalSubkey      = []byte{1}
	transferGenesisNumSubkey           = []byte{2}
)

// TransferGenesisSetMembershipKey returns the store key to check the presence of a transfer genesis transfer by its index
func TransferGenesisSetMembershipKey(
	rollappID string,
	index uint64,
) []byte {
	var key []byte
	// build the key bytes
	rollappIdBytes := []byte(rollappID)
	ixBytes := sdk.Uint64ToBigEndian(index)
	// concatenate the byte slices directly
	key = append(key, transferGenesisSetMembershipSubkey...)
	key = append(key, []byte("/")...)
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)
	key = append(key, ixBytes...)

	return key
}

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
