package cache_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/cache"
)

// Test struct to use in the cache
type testStruct struct {
	ID   int
	Name string
}

// Key extraction function for testStruct
func keyFunc(val testStruct) int {
	return val.ID
}

func TestInsertionOrdered_Upsert(t *testing.T) {
	testCases := []struct {
		name           string
		initial        []testStruct
		upserts        []testStruct
		expectedValues []testStruct
	}{
		{
			name:           "Insert single item",
			initial:        []testStruct{},
			upserts:        []testStruct{{ID: 1, Name: "Item 1"}},
			expectedValues: []testStruct{{ID: 1, Name: "Item 1"}},
		},
		{
			name:           "Insert multiple items",
			initial:        []testStruct{},
			upserts:        []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
			expectedValues: []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
		},
		{
			name:           "Update existing item",
			initial:        []testStruct{{ID: 1, Name: "Item 1"}},
			upserts:        []testStruct{{ID: 1, Name: "Updated Item 1"}},
			expectedValues: []testStruct{{ID: 1, Name: "Updated Item 1"}},
		},
		{
			name:           "Insert and update items",
			initial:        []testStruct{{ID: 1, Name: "Item 1"}},
			upserts:        []testStruct{{ID: 2, Name: "Item 2"}, {ID: 1, Name: "Updated Item 1"}, {ID: 0, Name: "Item 0"}},
			expectedValues: []testStruct{{ID: 1, Name: "Updated Item 1"}, {ID: 2, Name: "Item 2"}, {ID: 0, Name: "Item 0"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cache.NewInsertionOrdered(keyFunc, tc.initial...)
			c.Upsert(tc.upserts...)

			// Validate that the cache contains the expected values in the correct order
			require.Equal(t, tc.expectedValues, c.GetAll())
		})
	}
}

func TestInsertionOrdered_Get(t *testing.T) {
	testCases := []struct {
		name        string
		initial     []testStruct
		getID       int
		expectedVal testStruct
		found       bool
	}{
		{
			name:        "Get existing item",
			initial:     []testStruct{{ID: 1, Name: "Item 1"}},
			getID:       1,
			expectedVal: testStruct{ID: 1, Name: "Item 1"},
			found:       true,
		},
		{
			name:        "Get non-existing item",
			initial:     []testStruct{{ID: 1, Name: "Item 1"}},
			getID:       2,
			expectedVal: testStruct{},
			found:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cache.NewInsertionOrdered(keyFunc, tc.initial...)
			val, found := c.Get(tc.getID)

			require.Equal(t, tc.found, found)
			require.Equal(t, tc.expectedVal, val)
		})
	}
}

func TestInsertionOrdered_GetAll(t *testing.T) {
	testCases := []struct {
		name           string
		initial        []testStruct
		expectedValues []testStruct
	}{
		{
			name:           "Get all from empty cache",
			initial:        []testStruct{},
			expectedValues: []testStruct{},
		},
		{
			name:           "Get all from non-empty cache",
			initial:        []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
			expectedValues: []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cache.NewInsertionOrdered(keyFunc, tc.initial...)
			allValues := c.GetAll()

			require.Equal(t, tc.expectedValues, allValues)
		})
	}
}

func TestInsertionOrdered_Range(t *testing.T) {
	testCases := []struct {
		name           string
		initial        []testStruct
		stopID         int
		expectedValues []testStruct
	}{
		{
			name:           "Range over all values",
			initial:        []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
			stopID:         -1,
			expectedValues: []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
		},
		{
			name:           "Stop at specific value",
			initial:        []testStruct{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}},
			stopID:         1,
			expectedValues: []testStruct{{ID: 1, Name: "Item 1"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := cache.NewInsertionOrdered(keyFunc, tc.initial...)
			var collected []testStruct

			c.Range(func(v testStruct) bool {
				collected = append(collected, v)
				return v.ID == tc.stopID
			})

			require.Equal(t, tc.expectedValues, collected)
		})
	}
}
