package utils

import (
	"regexp"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/ethereum/go-ethereum/common"
)

func IsValidBech32AccountAddress(address string, matchAccountAddressBech32Prefix bool) bool {
	hrp, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return false
	}

	bytesCount := len(bz)
	if bytesCount != 20 && bytesCount != 32 /*32 bytes is interchain account*/ {
		return false
	}

	return !matchAccountAddressBech32Prefix || hrp == params.AccountAddressPrefix
}

var pattern0xHex = regexp.MustCompile(`^0x[a-f\d]+$`)

func IsValidHexAddress(address string) bool {
	length := len(address)
	if length != 42 && length != 66 /*32 bytes is interchain account*/ {
		return false
	}

	address = strings.ToLower(address)

	return pattern0xHex.MatchString(address)
}

func GetBytesFromHexAddress(address string) []byte {
	if !IsValidHexAddress(address) {
		panic("invalid hex address")
	}

	if len(address) == 66 {
		return common.HexToHash(address).Bytes()
	}

	return common.HexToAddress(address).Bytes()
}

func GetHexAddressFromBytes(bytes []byte) string {
	if len(bytes) == 32 {
		return strings.ToLower(common.BytesToHash(bytes).Hex())
	} else if len(bytes) == 20 {
		return strings.ToLower(common.BytesToAddress(bytes).Hex())
	} else {
		panic("invalid bytes length")
	}
}
