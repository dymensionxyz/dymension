package uptr

// To returns a pointer to the passed argument
func To[T any](x T) *T {
	return &x
}
