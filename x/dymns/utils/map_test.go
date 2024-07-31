package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSortedStringKeys(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		require.Empty(t, GetSortedStringKeys(map[string]bool{}))
		require.Empty(t, GetSortedStringKeys(map[string]string{}))
		require.Empty(t, GetSortedStringKeys(map[string]int{}))
		var nilMap map[string]bool
		require.Empty(t, GetSortedStringKeys(nilMap))
	})

	t.Run("sorted keys", func(t *testing.T) {
		require.Equal(t, []string{"a", "b", "c", "d"}, GetSortedStringKeys(map[string]bool{
			"b": true,
			"a": true,
			"c": true,
			"d": false,
		}))
		require.Equal(t, []string{"a", "b", "c", "d"}, GetSortedStringKeys(map[string]int64{
			"b": -1,
			"a": 1,
			"c": 2,
			"d": 0,
		}))
		require.Equal(t, []string{"a", "b", "c", "d"}, GetSortedStringKeys(map[string]any{
			"b": nil,
			"a": nil,
			"c": 2,
			"d": "f",
		}))
	})
}
