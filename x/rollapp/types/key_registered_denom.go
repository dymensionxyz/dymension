package types

const (
	// KeyRegisteredDenomPrefix is the prefix to retrieve all RegisteredDenom
	KeyRegisteredDenomPrefix = "RegisteredDenom/value/"
	// KeyRegisteredDenomSeparator is a null byte separator
	KeyRegisteredDenomSeparator = 0x00
)

// KeyRegisteredDenom returns the store key to retrieve a RegisteredDenom from the index fields
func KeyRegisteredDenom(rollappID, denom string) []byte {
	var key []byte
	key = append(key, []byte(rollappID)...)
	key = append(key, KeyRegisteredDenomSeparator)
	key = append(key, []byte(denom)...)
	return key
}

func RegisteredDenomPrefix(rollappID string) []byte {
	var prefix []byte
	prefix = append(prefix, []byte(rollappID)...)
	prefix = append(prefix, KeyRegisteredDenomSeparator)
	return prefix
}
