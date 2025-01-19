package ukeys

import (
	"bytes"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// EncodeTimeToKey combines a given byte slice (queueKey) with an encoded representation of a specified time (endTime).
// The resulting byte slice can be used for lexicographical sorting, as the encoded time is in big-endian order.
//
// The function ensures that different times will be sorted correctly when iterated using big-endian order,
// as the time encoding preserves the natural chronological order.
//
// Example:
//
//	queueKey := []byte("exampleKey")
//	endTime1 := time.Date(2023, time.October, 1, 12, 0, 0, 0, time.UTC)
//	endTime2 := time.Date(2023, time.October, 2, 12, 0, 0, 0, time.UTC)
//	encodedKey1 := EncodeTimeToKey(queueKey, endTime1)
//	encodedKey2 := EncodeTimeToKey(queueKey, endTime2)
//	fmt.Printf("%x\n", encodedKey1) // Output will be the hexadecimal representation of the combined byte slice for endTime1
//	fmt.Printf("%x\n", encodedKey2) // Output will be the hexadecimal representation of the combined byte slice for endTime2
//	fmt.Println(string(encodedKey1) < string(encodedKey2)) // This will print 'true' as endTime1 is before endTime2
func EncodeTimeToKey(queueKey []byte, endTime time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(endTime)
	prefixL := len(queueKey)
	bz := make([]byte, prefixL+len(timeBz))

	// copy the prefix
	copy(bz[:prefixL], queueKey)
	// copy the encoded time bytes
	copy(bz[prefixL:], timeBz)
	return bz
}

// GetTimeKey returns the time key used when getting a set of streams.
func GetTimeKey(timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(types.KeyPrefixTimestamp)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], types.KeyPrefixTimestamp)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

// FindIndex takes an array of IDs. Then return the index of a specific ID.
func FindIndex(IDs []uint64, ID uint64) int {
	for index, id := range IDs {
		if id == ID {
			return index
		}
	}
	return -1
}

// RemoveValue takes an array of IDs. Then finds the index of the IDs and remove those IDs from the array.
func RemoveValue(IDs []uint64, ID uint64) ([]uint64, int) {
	index := FindIndex(IDs, ID)
	if index < 0 {
		return IDs, index
	}
	IDs[index] = IDs[len(IDs)-1] // set last element to index
	return IDs[:len(IDs)-1], index
}

// CombineKeys combine bytes array into a single bytes.
func CombineKeys(keys ...[]byte) []byte {
	return bytes.Join(keys, types.KeyIndexSeparator)
}
