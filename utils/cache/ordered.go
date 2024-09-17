package cache

import "fmt"

// InsertionOrdered is a cache that preserves the insertion order of its elements.
// It maps keys of type K to values of type V and ensures that elements are iterated
// in the order they were inserted. This implementation is NOT thread-safe.
type InsertionOrdered[K comparable, V any] struct {
	key        func(V) K // A function that generates the key from a value
	nextIdx    int       // The index to assign to the next inserted element
	keyToIdx   map[K]int // Maps keys to their index in the idxToValue slice
	idxToValue []V       // Stores values in the order they were inserted
}

// NewInsertionOrdered creates and returns a new InsertionOrdered cache.
// It accepts a key function, which extracts a key of type K from a value of type V.
// Optionally, you can pass initial values to be inserted into the cache.
func NewInsertionOrdered[K comparable, V any](key func(V) K, initial ...V) *InsertionOrdered[K, V] {
	cache := &InsertionOrdered[K, V]{
		key:        key,
		nextIdx:    0,
		keyToIdx:   make(map[K]int, len(initial)),
		idxToValue: make([]V, 0, len(initial)),
	}
	// Insert the initial values (if any) into the cache
	cache.Upsert(initial...)
	return cache
}

// Upsert inserts or updates one or more values in the cache.
// If a value with the same key already exists, it will be updated.
// If the key is new, the value will be appended while preserving the insertion order.
func (c *InsertionOrdered[K, V]) Upsert(values ...V) {
	for _, value := range values {
		c.upsert(value)
	}
}

// upsert is an internal helper method that inserts or updates a single value.
// It extracts the key from the value, checks if it already exists in the cache,
// and updates the value if found. If not, it appends the new value to the cache.
func (c *InsertionOrdered[K, V]) upsert(value V) {
	key := c.key(value)
	idx, ok := c.keyToIdx[key]
	if ok {
		// If the key already exists, update the value
		c.idxToValue[idx] = value
	} else {
		// If the key does not exist, add a new entry
		idx = c.nextIdx
		c.nextIdx++
		c.keyToIdx[key] = idx
		c.idxToValue = append(c.idxToValue, value)
	}
}

// Get retrieves a value from the cache by its key.
// It returns the value and a boolean indicating whether the key was found.
// If the key does not exist, it returns the zero value of type V and false.
func (c *InsertionOrdered[K, V]) Get(key K) (zero V, found bool) {
	idx, ok := c.keyToIdx[key]
	if ok {
		return c.idxToValue[idx], true
	}
	return zero, false
}

// MustGet is Get that panics when the key is not found.
func (c *InsertionOrdered[K, V]) MustGet(key K) (zero V) {
	idx, ok := c.keyToIdx[key]
	if ok {
		return c.idxToValue[idx]
	}
	panic(fmt.Errorf("internal contract error: key is not found in the cache: %v", key))
}

// GetAll returns all values currently stored in the cache in their insertion order.
// This allows you to retrieve all values while preserving the order in which they were added.
func (c *InsertionOrdered[K, V]) GetAll() []V {
	return c.idxToValue
}

// Range iterates over the values in the cache in their insertion order.
// The provided function f is called for each value. If f returns true, the iteration stops early.
// This method allows for efficient traversal without needing to copy the entire cache.
func (c *InsertionOrdered[K, V]) Range(f func(V) bool) {
	for _, value := range c.idxToValue {
		stop := f(value)
		if stop {
			return
		}
	}
}
