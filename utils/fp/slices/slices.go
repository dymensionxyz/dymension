package slices

func Map[E, V any](slice []E, f func(E) V) []V {
	res := make([]V, 0, len(slice))
	for _, e := range slice {
		res = append(res, f(e))
	}
	return res
}

func StringToType[T ~string | ~[]byte](s string) T { return T(s) }

func TypeToString[T ~string | ~[]byte](t T) string { return string(t) }
