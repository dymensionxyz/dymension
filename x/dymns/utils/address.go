package utils

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/common"
)

var accAddrBech32Prefix = sdk.GetConfig().GetBech32AccountAddrPrefix()

// IsValidBech32AccountAddress returns true if the given string is a valid bech32 account address.
// Depends on the flag, it will check the prefix of the bech32 account address
// matches with host chain's account address prefix.
func IsValidBech32AccountAddress(address string, matchAccountAddressBech32Prefix bool) bool {
	hrp, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return false
	}

	bytesCount := len(bz)
	if bytesCount != 20 && bytesCount != 32 /*32 bytes is interchain account*/ {
		return false
	}

	return !matchAccountAddressBech32Prefix || hrp == accAddrBech32Prefix
}

// pattern0xHex is a regex pattern for 0x prefixed hex string.
// Length check is omitted.
var pattern0xHex = regexp.MustCompile(`^0x[a-f\d]+$`)

// IsValidHexAddress returns true if the given string is a valid hex address.
// It checks the length and the pattern of the hex address.
// The hex address can be either 20 bytes presents for normal accounts or 32 bytes presents for Interchain Accounts.
func IsValidHexAddress(address string) bool {
	length := len(address)
	if length != 42 && length != 66 /*32 bytes is interchain account*/ {
		return false
	}

	address = strings.ToLower(address)

	return pattern0xHex.MatchString(address)
}

// GetBytesFromHexAddress returns the bytes from the given hex address.
func GetBytesFromHexAddress(address string) []byte {
	if !IsValidHexAddress(address) {
		panic("invalid hex address")
	}

	if len(address) == 66 {
		return common.HexToHash(address).Bytes()
	}

	return common.HexToAddress(address).Bytes()
}

// GetHexAddressFromBytes returns the hex address from the given bytes.
func GetHexAddressFromBytes(bytes []byte) string {
	if len(bytes) == 32 {
		return strings.ToLower(common.BytesToHash(bytes).Hex())
	} else if len(bytes) == 20 {
		return strings.ToLower(common.BytesToAddress(bytes).Hex())
	} else {
		panic("invalid bytes length")
	}
}

// PossibleAccountRegardlessChain returns true if the given string is a POSSIBLE account address regardless of chain.
// There is no guarantee that the address is valid on the chain.
func PossibleAccountRegardlessChain(address string) bool {
	if length := len(address); length < 5 || length > 130 {
		return false
	}

	for _, r := range address {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' || r == ':' || r == '.' {
			continue
		}
		if r == '*' || r == '@' {
			// Stellar Federated Address
			continue
		}

		return false
	}

	return true
}
