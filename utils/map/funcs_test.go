package utilsmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedKeys(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		m := map[int]struct{}{}
		m[0] = struct{}{}
		m[2] = struct{}{}
		m[1] = struct{}{}
		res := SortedKeys(m)
		for i := 0; i < len(m); i++ {
			require.Equal(t, i, res[i])
		}
	})
}
