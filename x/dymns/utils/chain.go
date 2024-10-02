package utils

import (
	"regexp"
	"strings"

	cometbfttypes "github.com/cometbft/cometbft/types"
)

// patternValidChainId is a regex pattern for valid chain id format.
var patternValidChainId = regexp.MustCompile(`^[a-z]+(-[a-z]+)?(_\d+)?(-\d+)?$`)

// IsValidChainIdFormat returns true if the given string is a valid chain id format.
// It checks the length and the pattern of the chain id.
// The chain id length can be between 3 and 50 characters.
func IsValidChainIdFormat(chainId string) bool {
	// TODO: move validation functions like this to sdk-utils repo
	if len(chainId) < 3 || len(chainId) > cometbfttypes.MaxChainIDLen {
		return false
	}

	return patternValidChainId.MatchString(chainId)
}

// patternNumericOnly is a regex pattern for numeric only.
var patternNumericOnly = regexp.MustCompile(`^\d+$`)

// IsValidEIP155ChainId returns true if the given string is a valid EIP155 chain id format.
// Format should be positive numeric only, except zero.
func IsValidEIP155ChainId(eip155ChainId string) bool {
	if !patternNumericOnly.MatchString(eip155ChainId) {
		return false
	}

	if strings.HasPrefix(eip155ChainId, "0") {
		// Case 1: prevent zero, used to protect output from failed-to-parse cases
		// Case 2: prevent leading zeros
		return false
	}

	return true
}
