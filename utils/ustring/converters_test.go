package ustring_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/ustring"
)

func TestStringToType(t *testing.T) {
	type customString string
	type customBytes []byte

	testCases := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "Convert string to customString",
			input:    "hello",
			expected: customString("hello"),
		},
		{
			name:     "Convert string to customBytes",
			input:    "hello",
			expected: customBytes("hello"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.expected.(type) {
			case customString:
				result := ustring.FromString[customString](tc.input)
				require.Equal(t, tc.expected, result)
			case customBytes:
				result := ustring.FromString[customBytes](tc.input)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestTypeToString(t *testing.T) {
	type customString string
	type customBytes []byte

	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Convert customString to string",
			input:    customString("hello"),
			expected: "hello",
		},
		{
			name:     "Convert customBytes to string",
			input:    customBytes("hello"),
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch input := tc.input.(type) {
			case customString:
				result := ustring.ToString(input)
				require.Equal(t, tc.expected, result)
			case customBytes:
				result := ustring.ToString(input)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}
