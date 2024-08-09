package utils

import "regexp"

// patternValidChainId is a regex pattern for valid chain id format.
var patternValidChainId = regexp.MustCompile(`^[a-z]+(-[a-z]+)?(_\d+)?(-\d+)?$`)

// IsValidChainIdFormat returns true if the given string is a valid chain id format.
// It checks the length and the pattern of the chain id.
// The chain id length can be between 3 and 47 characters.
func IsValidChainIdFormat(chainId string) bool {
	if len(chainId) < 3 || len(chainId) > 47 {
		return false
	}

	return patternValidChainId.MatchString(chainId)
}
