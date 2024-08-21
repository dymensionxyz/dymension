package utils

import "regexp"

// patternValidateAlias is a regex pattern for validating Alias (partially).
var patternValidateAlias = regexp.MustCompile(`^[a-z\d]{1,32}$`)

// IsValidAlias returns true if the given string is a valid Alias.
func IsValidAlias(alias string) bool {
	if alias == "" {
		return false
	}

	if len(alias) > MaxAliasLength {
		return false
	}

	return patternValidateAlias.MatchString(alias)
}
