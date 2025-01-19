package uptr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTo(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type X struct {
			x int
		}

		x := X{42}

		require.Equal(t, x.x, (*To(x)).x)
	})
	t.Run("r value", func(t *testing.T) {
		x := 42
		f := func() int {
			return x
		}

		require.Equal(t, x, *To(f()))
	})
}
