package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ binary.ByteOrder

const (
	// StateInfoKeyPrefix is the prefix to retrieve all StateInfo
	StateInfoKeyPrefix          = "StateInfo/value/"
	StateInfoIndexKeyPartLength = 8 + 1 + 1 // BigEndian + "/" + "/"
)

// StateInfoKey returns the store key to retrieve a StateInfo from the index fields
func StateInfoKey(
	stateInfoIndex StateInfoIndex,
) []byte {
	var key []byte
	// build the key bytes
	rollappIdBytes := []byte(stateInfoIndex.RollappId)
	stateInfoIndexBytes := sdk.Uint64ToBigEndian(stateInfoIndex.Index)
	// concatenate the byte slices directly
	key = append(key, rollappIdBytes...)
	key = append(key, []byte("/")...)
	key = append(key, stateInfoIndexBytes...)
	key = append(key, []byte("/")...)

	return key
}

// StateInfoIndexFromKey returns the StateInfoIndex from a store key.
// The value of StateInfoIndexKeyPartLength will always be shorter than the key itself,
// because the key contains the rollappId and the BigEndian representation of the index,
// which is always 8 bytes long.
func StateInfoIndexFromKey(key []byte) StateInfoIndex {
	l := len(key)
	rollappId := string(key[:l-StateInfoIndexKeyPartLength])
	return StateInfoIndex{
		RollappId: rollappId,
		Index:     sdk.BigEndianToUint64(key[len(rollappId)+1 : l-1]),
	}
}

// StateInfoIndexKeyFromTimestampKey returns the StateInfoIndex key from a timestamp key by removing the timestamp prefix.
func StateInfoIndexKeyFromTimestampKey(keyTS []byte) []byte {
	return keyTS[TimestampPrefixLen:] // remove the timestamp prefix
}
