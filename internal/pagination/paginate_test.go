package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/internal/pagination"
)

type testIterator struct {
	data  []int
	index int
}

func newTestIterator(data []int) *testIterator {
	return &testIterator{data: data, index: 0}
}

func (t *testIterator) Next() {
	t.index++
}

func (t *testIterator) Value() int {
	return t.data[t.index]
}

func (t *testIterator) Valid() bool {
	return t.index < len(t.data)
}

func TestPaginate(t *testing.T) {
	testCases := []struct {
		name      string
		iterator  pagination.Iterator[int]
		perPage   uint64
		stopValue int
		expected  uint64
	}{
		{
			name:      "Empty iterator",
			iterator:  newTestIterator([]int{}),
			perPage:   5,
			stopValue: -1,
			expected:  0,
		},
		{
			name:      "Non-Empty iterator less than perPage",
			iterator:  newTestIterator([]int{1, 2, 3}),
			perPage:   10,
			stopValue: -1,
			expected:  3,
		},
		{
			name:      "Non-empty iterator greater than perPage",
			iterator:  newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			perPage:   5,
			stopValue: -1,
			expected:  5,
		},
		{
			name:      "Zero perPage",
			iterator:  newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			perPage:   0,
			stopValue: 6,
			expected:  0,
		},
		{
			name:      "Non-Empty iterator with stop condition",
			iterator:  newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			perPage:   10,
			stopValue: 3,
			expected:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pagination.Paginate(tc.iterator, tc.perPage, func(i int) pagination.Stop { return i == tc.stopValue })
			require.Equal(t, tc.expected, result)
		})
	}
}
