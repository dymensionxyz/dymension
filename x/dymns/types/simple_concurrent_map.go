package types

import "sync"

// SimpleConcurrentMap is a simple implementation of concurrent map.
type SimpleConcurrentMap[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

func NewSimpleConcurrentMap[K comparable, V any]() *SimpleConcurrentMap[K, V] {
	return &SimpleConcurrentMap[K, V]{
		m: make(map[K]V),
	}
}

func (m *SimpleConcurrentMap[K, V]) Set(key K, value V) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *SimpleConcurrentMap[K, V]) Has(key K) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.m[key]
	return ok
}

func (m *SimpleConcurrentMap[K, V]) Get(key K) (v V, found bool) {
	m.RLock()
	defer m.RUnlock()
	value, ok := m.m[key]
	return value, ok
}

func (m *SimpleConcurrentMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
}

func (m *SimpleConcurrentMap[K, V]) Clear() {
	m.Lock()
	defer m.Unlock()
	m.m = make(map[K]V)
}
