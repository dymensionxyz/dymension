package ustring

// FromString converts a string to a custom type T, where T can be either a string or a byte slice.
// May be used as a parameter for uslice.Map.
func FromString[T ~string | ~[]byte](s string) T { return T(s) }

// ToString converts a custom type T, where T can be either a string or a byte slice, back to a string.
// May be used as a parameter for uslice.Map.
func ToString[T ~string | ~[]byte](t T) string { return string(t) }
