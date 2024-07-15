package cli_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []types.GaugeWeight
		expectedErr error
	}{
		{
			name:  "Valid input",
			input: "15:60,10:70,12:10",
			expected: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			expectedErr: false,
		},
		{
			name:        "Invalid input",
			input:       "15,10:70,12:10",
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "Value out of range",
			input:       "15:101,10:70,12:10",
			expected:    nil,
			expectedErr: true,
		},
		{
			name:        "Empty input",
			input:       "",
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cli.ParseGaugeWeights(tt.input)
			if (err != nil) != tt.expectedErr {
				t.Errorf("parse() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			require.Equal(t, tt.expected, got)
		})
	}
}
