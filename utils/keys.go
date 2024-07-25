package utils

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
