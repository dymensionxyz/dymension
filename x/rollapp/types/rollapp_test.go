package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRollappGetRevisionForHeight(t *testing.T) {
	rollapp := Rollapp{Revisions: []Revision{{Number: 1, StartHeight: 10}}}

	revision, found := rollapp.GetRevisionForHeight(9)
	require.False(t, found)
	require.Equal(t, Revision{}, revision)
	require.False(t, rollapp.IsRevisionStartHeight(0, 0))

	revision, found = rollapp.GetRevisionForHeight(10)
	require.True(t, found)
	require.Equal(t, Revision{Number: 1, StartHeight: 10}, revision)
	require.True(t, rollapp.IsRevisionStartHeight(1, 10))
}
