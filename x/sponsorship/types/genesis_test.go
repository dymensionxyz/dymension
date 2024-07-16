package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func TestGenesis(t *testing.T) {
	tests := []struct {
		name          string
		input         *types.GenesisState
		errorIs       error
		errorContains string
	}{
		{
			name: "Valid input",
			input: &types.GenesisState{
				Params: types.Params{
					MinAllocationWeight: math.NewInt(20),
					MinVotingPower:      math.NewInt(20),
				},
			},
			errorIs:       nil,
			errorContains: "",
		},
		{
			name:          "Default is valid",
			input:         types.DefaultGenesis(),
			errorIs:       nil,
			errorContains: "",
		},
		{
			name: "Invalid params, MinAllocationWeight < 0",
			input: &types.GenesisState{
				Params: types.Params{
					MinAllocationWeight: math.NewInt(-20),
					MinVotingPower:      math.NewInt(20),
				},
			},
			errorIs:       types.ErrInvalidGenesis,
			errorContains: "MinAllocationWeight must be >= 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

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
