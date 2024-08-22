package utils

import (
	"regexp"
	"strings"
)

// patternValidateDymNameStep1 is a regex pattern for validating Dym-Name in first step.
var patternValidateDymNameStep1 = regexp.MustCompile(`^[a-z\d]+([a-z\d_-]*[a-z\d]+)?$`)

// IsValidDymName returns true if the given string is a valid Dym-Name.
// Read code and the comments for more details.
func IsValidDymName(dymName string) bool {
	if len(dymName) > MaxDymNameLength {
		return false
	}

	if dymName == "" {
		return false
	}

	// step 1: check if the dym name is valid with following rules
	// 1. only lowercase letters, digits, hyphens, and underscores are allowed
	// 2. the first character must be a letter or a digit
	// 3. the last character must be a letter or a digit

	if !patternValidateDymNameStep1.MatchString(dymName) {
		return false
	}

	// step 2: check if the dym name does not contain consecutive hyphens or underscores
	for i := 0; i < len(dymName)-1; i++ {
		if (dymName[i] == '-' || dymName[i] == '_') && (dymName[i+1] == '-' || dymName[i+1] == '_') {
			return false
		}
	}

	return true
}

// IsValidSubDymName returns true if the given string is a valid Sub-Name of Dym-Name.
func IsValidSubDymName(subDymName string) bool {
	if subDymName == "" {
		// allowed to be empty, means no sub name
		return true
	}

	if len(subDymName) > MaxSubNameLength {
		return false
	}

	if strings.HasPrefix(subDymName, ".") || strings.HasSuffix(subDymName, ".") {
		return false
	}

	spl := strings.Split(subDymName, ".")
	for _, s := range spl {
		if s == "" {
			return false
		}

		if !IsValidDymName(s) {
			return false
		}
	}

	return true
}
