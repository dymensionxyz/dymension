package utils

import "slices"

// GetSortedStringKeys returns the sorted keys of a map[string]V.
func GetSortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
