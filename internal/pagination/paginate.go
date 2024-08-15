package pagination

type Iterator[T any] interface {
	Next()
	Value() T
	Valid() bool
}

type Stop bool

const Break Stop = true
const Continue Stop = false

// Paginate is a function that paginates over an iterator. The callback is executed for each iteration and if it
// returns true, the pagination stops. The function returns the amount of iterations before stopping.
func Paginate[T any](iter Iterator[T], perPage uint64, cb func(T) Stop) uint64 {
	iterations := uint64(0)
	for ; iterations < perPage && iter.Valid(); iter.Next() {
		iterations++

		stop := cb(iter.Value())
		if stop {
			break
		}
	}
	return iterations
}
