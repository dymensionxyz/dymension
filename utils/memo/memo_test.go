package utilsmemo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		memo := `{
"foo" : 0,
"bar" : 1,
}`

		res, err := Merge(memo, map[string]int{"fizz": 2, "buzz": 3})
		require.NoError(t, err)
		require.True(t, strings.Contains(res, "fizz"))
	})
	t.Run("conflict", func(t *testing.T) {
	})
}
