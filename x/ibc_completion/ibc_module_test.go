package ibc_completion

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoHasConflictingMiddleware(t *testing.T) {
	memo := map[string]interface{}{
		"forward": map[string]interface{}{},
	}
	memoBz, err := json.Marshal(memo)
	require.NoError(t, err)
	require.True(t, memoHasConflictingMiddleware(memoBz))

	// test non-conflicting memo
	memo = map[string]interface{}{
		"other": "value",
	}
	memoBz, err = json.Marshal(memo)
	require.NoError(t, err)
	require.False(t, memoHasConflictingMiddleware(memoBz))
}
