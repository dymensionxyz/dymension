package uslice

// ToKeySet converts a slice of elements into a set (represented as a map with empty structs).
// This function ensures that all elements in the resulting map are unique.
func ToKeySet[E comparable](slice []E) map[E]struct{} {
	res := make(map[E]struct{}, len(slice))
	for _, e := range slice {
		res[e] = struct{}{}
	}
	return res
}

// Map applies a provided function f to each element in the input slice and returns a slice of modified values.
func Map[E, V any](slice []E, f func(E) V) []V {
	res := make([]V, 0, len(slice))
	for _, e := range slice {
		res = append(res, f(e))
	}
	return res
}
