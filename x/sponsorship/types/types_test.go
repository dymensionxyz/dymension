package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestValidateGaugeWeights(t *testing.T) {
	tests := []struct {
		name          string
		input         []types.GaugeWeight
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid input",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Weight > 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(101)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "weight must be <= 100",
		},
		{
			name: "Weight < 0",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(-40)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "weight must be > 0",
		},
		{
			name: "Weight == 0",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(0)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "weight must be > 0",
		},
		{
			name: "Sum of weighs > 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(30)},
				{GaugeId: 12, Weight: math.NewInt(20)},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "total weight must equal 100",
		},
		{
			name: "Sum of weighs < 100",
			input: []types.GaugeWeight{
				{GaugeId: 15, Weight: math.NewInt(60)},
				{GaugeId: 10, Weight: math.NewInt(20)},
				{GaugeId: 12, Weight: math.NewInt(10)},
			},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "total weight must equal 100",
		},
		{
			name:          "Empty input",
			input:         []types.GaugeWeight{},
			errorIs:       types.ErrInvalidDistribution,
			errorContains: "total weight must equal 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateGaugeWeights(tt.input)

			expectError := tt.errorIs != nil
			switch expectError {
			case true:
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errorIs)
				require.Contains(t, err.Error(), tt.errorContains)
			case false:
				require.NoError(t, err)
			}
		})
	}
}
