package utils

import "regexp"

// patternValidateServiceKey is a regex pattern for validating a service record key
// of a DCT_SERVICE Dym-Name config: lowercase alphanumeric, may contain hyphens,
// starting with an alphanumeric, up to 32 characters.
var patternValidateServiceKey = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

// IsValidServiceKey returns true if the given string is a valid service record key.
func IsValidServiceKey(serviceKey string) bool {
	return patternValidateServiceKey.MatchString(serviceKey)
}
