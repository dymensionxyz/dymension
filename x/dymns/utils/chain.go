package utils

import "regexp"

var patternValidChainId = regexp.MustCompile(`^[a-z]+(-[a-z]+)?(_\d+)?(-\d+)?$`)

func IsValidChainIdFormat(chainId string) bool {
	if len(chainId) < 3 || len(chainId) > 47 {
		return false
	}

	return patternValidChainId.MatchString(chainId)
}
