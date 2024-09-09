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
			input: "15=60000000000000000000,10=3000,12=10000",
			expected: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(60)},
				{GaugeId: 10, Weight: math.NewInt(3000)},
				{GaugeId: 12, Weight: math.NewInt(10000)},
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name:  "Valid, sum of weighs < 100",
			input: "15=30000000000000000000,10=30000000000000000000,12=39000000000000000000",
			expected: []types.GaugeWeight{
				{GaugeId: 15, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 10, Weight: types.DYM.MulRaw(30)},
				{GaugeId: 12, Weight: types.DYM.MulRaw(39)},
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "Invalid input format",
			input:         "15,10=70000000000000000000,12=10000000000000000000",
			expected:      nil,
			expectError:   true,
			errorContains: "invalid gauge weight format",
		},
		{
			name:          "Weight > 100",
			input:         "15=101000000000000000000,10=70000000000000000000,12=10000000000000000000",
			expected:      nil,
			expectError:   true,
			errorContains: "weight must be between 1 and 100 * 10^18, got 101000000000000000000",
		},
		{
			name:          "Weight < 0",
			input:         "15=-10,10=70,12=10",
			expected:      nil,
			expectError:   true,
			errorContains: "weight must be between 1 and 100 * 10^18, got -10",
		},
		{
			name:          "Sum of weighs > 100",
			input:         "15=30000000000000000000,10=30000000000000000000,12=41000000000000000000",
			expected:      nil,
			expectError:   true,
			errorContains: "total weight must be less than 100 * 10^18, got 101000000000000000000",
		},
		{
			name:          "Zero weight",
			input:         "15=0,10=30000000000000000000,12=41000000000000000000",
			expected:      nil,
			expectError:   true,
			errorContains: "weight must be between 1 and 100 * 10^18, got 0",
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
