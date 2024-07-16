package cli_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestParseGaugeWeights(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      []types.GaugeWeight
		expectError   bool
		errorContains string
	}{
		{
			name:  "Valid input",
			input: "15:60,10:30,12:10",
			expected: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "Invalid input format",
			input:         "15,10:70,12:10",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid gauge weight format",
		},
		{
			name:          "Weight > 100",
			input:         "15:101,10:70,12:10",
			expected:      nil,
			expectError:   true,
			errorContains: "weight must be between 0 and 100",
		},
		{
			name:          "Weight < 0",
			input:         "15:-10,10:70,12:10",
			expected:      nil,
			expectError:   true,
			errorContains: "weight must be between 0 and 100",
		},
		{
			name:          "Sum of weighs > 100",
			input:         "15:30,10:30,12:41",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid gauge weights",
		},
		{
			name:          "Sum of weighs < 100",
			input:         "15:30,10:30,12:39",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid gauge weights",
		},
		{
			name:          "Empty input",
			input:         "",
			expected:      nil,
			expectError:   true,
			errorContains: "input weights must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := cli.ParseGaugeWeights(tt.input)

			switch tt.expectError {
			case true:
				require.Error(t, err)
				require.Nil(t, actual)
				require.Contains(t, err.Error(), tt.errorContains)
			case false:
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}
