package ucli

import "strings"

func Affirmative(s string) bool {
	l := strings.ToLower(s)
	truthy := map[string]struct{}{
		"true": {},
		"t":    {},
		"yes":  {},
		"y":    {},
	}
	_, ok := truthy[l]
	return ok
}
