package utils

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: move to sdk-utils (https://github.com/dymensionxyz/dymension/issues/1008)

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
