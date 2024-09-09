package cache_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/cache"
)

var (
	k1   = "1"
	k2   = "2"
	k3   = "3"
	k4   = "4"
	k1v1 = item{key: k1, value: "1"}
	k1v2 = item{key: k1, value: "2"}
	k2v1 = item{key: k2, value: "1"}
	k2v2 = item{key: k2, value: "2"}
	k3v1 = item{key: k3, value: "1"}
	k4v1 = item{key: k4, value: "1"}
)

func TestOrdered(t *testing.T) {
	t.Parallel()

	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		c := cache.NewInsertionOrdered(itemKey)

		c.Add(k3v1) // add one
		require.Equal(t, []item{k3v1}, c.GetAll())
		c.Reset()

		c.Add(k3v1, k2v1) // add two different
		require.Equal(t, []item{k3v1, k2v1}, c.GetAll())

		c.Add(k2v2) // add existent
		require.Equal(t, []item{k3v1, k2v2}, c.GetAll())

		c.Add(k3v1) // add existent
		require.Equal(t, []item{k2v2, k3v1}, c.GetAll())
	})

	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		c := cache.NewInsertionOrdered(itemKey)

		c.Add(k3v1, k1v1, k4v1, k2v1)
		require.Equal(t, []item{k3v1, k1v1, k4v1, k2v1}, c.GetAll())

		c.Delete(k1) // delete middle
		require.Equal(t, []item{k3v1, k4v1, k2v1}, c.GetAll())

		c.Delete(k2) // delete back
		require.Equal(t, []item{k3v1, k4v1}, c.GetAll())

		c.Add(k1v2, k4v1, k2v2)
		require.Equal(t, []item{k3v1, k1v2, k4v1, k2v2}, c.GetAll())

		c.Delete(k1, k2) // delete several
		require.Equal(t, []item{k3v1, k4v1}, c.GetAll())

		c.Delete(k3) // delete front
		require.Equal(t, []item{k4v1}, c.GetAll())

		c.Delete(k4) // delete last
		require.Equal(t, []item{}, c.GetAll())
	})
}

type item struct {
	key, value string
}

func itemKey(i item) string {
	return i.key
}
