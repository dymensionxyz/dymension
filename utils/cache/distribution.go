package cache

import sdk "github.com/cosmos/cosmos-sdk/types"

type Distribution[K comparable, V any] struct {
	c *InsertionOrdered[K, V]

	addCoins        func(V, sdk.Coins) V
	totalDistrCoins sdk.Coins
}

func NewDistribution[K comparable, V any](
	values []V,
	addCoins func(V, sdk.Coins) V,
	key func(V) K,
) *Distribution[K, V] {
	cache := NewInsertionOrdered(key)
	cache.Add(values...)
	return &Distribution[K, V]{
		c:               cache,
		addCoins:        addCoins,
		totalDistrCoins: sdk.NewCoins(),
	}
}

// AddValueWithCoins sets a key-value pair to the cache after adding coins to the value using the addCoinsFn function.
// If the key is present, update the existing value. The method keeps track of all added coins.
// Returns the updated value.
func (c *Distribution[K, V]) AddValueWithCoins(value V, coins sdk.Coins) V {
	value = c.addCoins(value, coins)
	c.c.Add(value)
	c.totalDistrCoins = c.totalDistrCoins.Add(coins...)
	return value
}

// Add adds a key-value pair to the cache. If the key is present, update the existing value.
func (c *Distribution[K, V]) Add(value V) {
	c.c.Add(value)
}

func (c *Distribution[K, V]) Get(key K) (V, bool) {
	return c.c.Get(key)
}

// GetValues returns all values stored in the cache pursuing the insertion order.
func (c *Distribution[K, V]) GetValues() []V {
	return c.c.GetAll()
}

func (c *Distribution[K, V]) TotalDistrCoins() sdk.Coins {
	return c.totalDistrCoins
}

func (c *Distribution[K, V]) Range(f func(V) bool) {
	c.c.Range(f)
}
