package uslice_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/dymensionxyz/dymension/v3/utils/uslice"
)

func TestToKeySet(t *testing.T) {
	testCases := []struct {
		name     string
		input    []int
		expected map[int]struct{}
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			expected: map[int]struct{}{},
		},
		{
			name:     "Single element slice",
			input:    []int{1},
			expected: map[int]struct{}{1: {}},
		},
		{
			name:     "Multiple elements, no duplicates",
			input:    []int{1, 2, 3},
			expected: map[int]struct{}{1: {}, 2: {}, 3: {}},
		},
		{
			name:     "Multiple elements with duplicates",
			input:    []int{1, 2, 2, 3, 1},
			expected: map[int]struct{}{1: {}, 2: {}, 3: {}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := uslice.ToKeySet(tc.input)
			require.True(t, maps.Equal(tc.expected, result))
		})
	}
}

func TestMap(t *testing.T) {
	testCases := []struct {
		name     string
		input    []int
		mapFunc  func(int) int
		expected []int
	}{
		{
			name:     "Empty slice",
			input:    []int{},
			mapFunc:  func(x int) int { return x * 2 },
			expected: []int{},
		},
		{
			name:     "Simple mapping",
			input:    []int{1, 2, 3},
			mapFunc:  func(x int) int { return x * 2 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "Identity mapping",
			input:    []int{1, 2, 3},
			mapFunc:  func(x int) int { return x },
			expected: []int{1, 2, 3},
		},
		{
			name:     "Mapping to constant",
			input:    []int{1, 2, 3},
			mapFunc:  func(int) int { return 42 },
			expected: []int{42, 42, 42},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := uslice.Map(tc.input, tc.mapFunc)
			require.Equal(t, tc.expected, result)
		})
	}
}
