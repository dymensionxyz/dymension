package types

import "cosmossdk.io/collections"

const (
	// KeyRegisteredDenomPrefix is the prefix to retrieve all RegisteredDenom
	KeyRegisteredDenomPrefix = "RegisteredDenom/value/"
	// KeyRegisteredDenomSeparator is a null byte separator
	KeyRegisteredDenomSeparator = 0x00
)

type RollappDenomKey collections.KeySet[string]

// KeyRegisteredDenom returns the store key to retrieve a RegisteredDenom from the index fields
func KeyRegisteredDenom(rollappID, denom string) []byte {
	return append(RegisteredDenomPrefix(rollappID), []byte(denom)...)
}

func RegisteredDenomPrefix(rollappID string) []byte {
	return append([]byte(rollappID), KeyRegisteredDenomSeparator)
}
