package pagination

type Iterator[T any] interface {
	Next()
	Value() T
	Valid() bool
}

// Paginate is a function that paginates over an iterator. The callback is executed for each iteration and if it
// returns true, the pagination stops. The callback also returns the number of operations performed during the call.
// That is, one iteration may be complex and thus return >1 operation num. For example, in case if the called decides
// that the iteration is heavy or time-consuming. Paginate also allows to specify the maximum number of operations
// that may be accumulated during the execution. If this number is exceeded, then Paginate exits.
// The function returns the total number of operations performed before stopping.
func Paginate[T any](
	iter Iterator[T],
	maxOperations uint64,
	cb func(T) (stop bool, operations uint64),
) uint64 {
	totalOperations := uint64(0)
	stop := false
	for ; !stop && totalOperations < maxOperations && iter.Valid(); iter.Next() {
		var operations uint64
		stop, operations = cb(iter.Value())
		totalOperations += operations
	}
	return totalOperations
}
