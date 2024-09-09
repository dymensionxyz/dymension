package cache

// InsertionOrdered cache preserves insertion-order.
// NOT thread safe!
type InsertionOrdered[K comparable, V any] struct {
	dataIdx map[K]*Node[V]
	key     func(V) K
	data    *List[V]
}

func NewInsertionOrdered[K comparable, V any](key func(V) K) *InsertionOrdered[K, V] {
	cache := &InsertionOrdered[K, V]{
		dataIdx: make(map[K]*Node[V], 0),
		key:     key,
		data:    NewList[V](),
	}
	return cache
}

func (c *InsertionOrdered[K, V]) Reset(values ...V) {
	c.dataIdx = make(map[K]*Node[V], len(values))
	c.data = NewList[V]()
	c.add(values...)
}

func (c *InsertionOrdered[K, V]) Add(values ...V) {
	c.add(values...)
}

func (c *InsertionOrdered[K, V]) add(values ...V) {
	for _, value := range values {
		key := c.key(value)
		if value, ok := c.dataIdx[key]; ok {
			value.Delete()
		}
		c.dataIdx[key] = c.data.Insert(value)
	}
}

func (c *InsertionOrdered[K, V]) Delete(keys ...K) {
	for _, key := range keys {
		if value, ok := c.dataIdx[key]; ok {
			value.Delete()
		}
		delete(c.dataIdx, key)
	}
}

func (c *InsertionOrdered[K, V]) Get(key K) (V, bool) {
	value, ok := c.dataIdx[key]
	if ok {
		return value.elem, true
	}
	var zero V
	return zero, false
}

func (c *InsertionOrdered[K, V]) GetAll() []V {
	res := make([]V, 0, len(c.dataIdx))
	c.data.Range(func(v V) bool {
		res = append(res, v)
		return true
	})
	return res
}

func (c *InsertionOrdered[K, V]) Filter(condition func(V) bool) []V {
	var res []V
	c.data.Range(func(v V) bool {
		if condition(v) {
			res = append(res, v)
		}
		return true
	})
	return res
}

func (c *InsertionOrdered[K, V]) FindFirst(condition func(V) bool) (V, bool) {
	var res V
	var found bool
	c.data.Range(func(v V) bool {
		if condition(v) {
			res = v
			found = true
			return false
		}
		return true
	})
	return res, found
}

func (c *InsertionOrdered[K, V]) Range(f func(V) bool) {
	c.data.Range(f)
}
