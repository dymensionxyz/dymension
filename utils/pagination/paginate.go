package pagination

type Iterator[T any] interface {
	Next()
	Value() T
	Valid() bool
}

// Paginate is a function that paginates over an iterator. The callback is executed for each iteration and if it
// returns true, the pagination stops. The callback also returns the weight of the iteration. That is, one iteration
// may be counted as multiple iterations. For example, in case if the called decides that the iteration is heavy
// or time-consuming. The function returns the amount of iterations before stopping.
func Paginate[T any](
	iter Iterator[T],
	perPage uint64,
	cb func(T) (stop bool, weight uint64),
) uint64 {
	iterations := uint64(0)
	stop := false
	for ; !stop && iterations < perPage && iter.Valid(); iter.Next() {
		var weight uint64
		stop, weight = cb(iter.Value())
		iterations += weight
	}
	return iterations
}
