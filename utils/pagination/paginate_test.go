package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/pagination"
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
		name               string
		iterator           pagination.Iterator[int]
		maxIterations      uint64
		stopValue          int
		expectedIterations uint64
		iterationWeight    uint64
	}{
		{
			name:               "Empty iterator",
			iterator:           newTestIterator([]int{}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 0,
			iterationWeight:    1,
		},
		{
			name:               "Non-Empty iterator less than maxIterations",
			iterator:           newTestIterator([]int{1, 2, 3}),
			maxIterations:      10,
			stopValue:          -1,
			expectedIterations: 3,
			iterationWeight:    1,
		},
		{
			name:               "Non-empty iterator greater than maxIterations",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 5,
			iterationWeight:    1,
		},
		{
			name:               "Zero maxIterations",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      0,
			stopValue:          6,
			expectedIterations: 0,
			iterationWeight:    1,
		},
		{
			name:               "Non-Empty iterator with stop condition",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      10,
			stopValue:          3,
			expectedIterations: 3,
			iterationWeight:    1,
		},
		{
			name:               "Empty iterator, >1 iteration weight",
			iterator:           newTestIterator([]int{}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 0,
			iterationWeight:    3,
		},
		{
			name:               "Non-Empty iterator less than maxIterations, >1 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3}),
			maxIterations:      10,
			stopValue:          -1,
			expectedIterations: 9,
			iterationWeight:    3,
		},
		{
			name:               "Non-empty iterator greater than maxIterations, >1 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 6,
			iterationWeight:    3,
		},
		{
			name:               "Zero maxIterations, >1 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      0,
			stopValue:          6,
			expectedIterations: 0,
			iterationWeight:    3,
		},
		{
			name:               "Non-Empty iterator with stop condition, >1 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      10,
			stopValue:          3,
			expectedIterations: 9,
			iterationWeight:    3,
		},
		{
			name:               "Empty iterator, 0 iteration weight",
			iterator:           newTestIterator([]int{}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 0,
			iterationWeight:    0,
		},
		{
			name:               "Non-Empty iterator less than maxIterations, 0 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3}),
			maxIterations:      10,
			stopValue:          -1,
			expectedIterations: 0,
			iterationWeight:    0,
		},
		{
			name:               "Non-empty iterator greater than maxIterations, 0 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      5,
			stopValue:          -1,
			expectedIterations: 0,
			iterationWeight:    0,
		},
		{
			name:               "Zero maxIterations, 0 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      0,
			stopValue:          6,
			expectedIterations: 0,
			iterationWeight:    0,
		},
		{
			name:               "Non-Empty iterator with stop condition, 0 iteration weight",
			iterator:           newTestIterator([]int{1, 2, 3, 4, 5, 6, 7}),
			maxIterations:      10,
			stopValue:          3,
			expectedIterations: 0,
			iterationWeight:    0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pagination.Paginate(tc.iterator, tc.maxIterations, func(i int) (bool, uint64) { return i == tc.stopValue, tc.iterationWeight })
			require.Equal(t, tc.expectedIterations, result)
		})
	}
}
