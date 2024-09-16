package slices

func ToKeySet[E comparable](slice []E) map[E]struct{} {
	res := make(map[E]struct{}, len(slice))
	for _, e := range slice {
		res[e] = struct{}{}
	}
	return res
}

func Merge[S ~[]V, V any](s ...S) S {
	var l int
	for _, slice := range s {
		l += len(slice)
	}
	r := make(S, 0, l)
	for _, slice := range s {
		r = append(r, slice...)
	}
	return r
}

func Map[E, V any](slice []E, f func(E) V) []V {
	res := make([]V, 0, len(slice))
	for _, e := range slice {
		res = append(res, f(e))
	}
	return res
}

func StringToType[T ~string | ~[]byte](s string) T { return T(s) }

func TypeToString[T ~string | ~[]byte](t T) string { return string(t) }
