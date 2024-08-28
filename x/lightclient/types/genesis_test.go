package types_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisValidate(t *testing.T) {
	tests := []struct {
		name  string
		g     types.GenesisState
		valid bool
	}{
		{
			name: "valid",
			g: types.GenesisState{
				CanonicalClients: []*types.CanonicalClient{
					{RollappId: "rollapp-1", IbcClientId: "client-1"},
					{RollappId: "rollapp-2", IbcClientId: "client-2"},
				},
				ConsensusStateSigners: []*types.ConsensusStateSigner{
					{IbcClientId: "client-1", Height: 1, Signer: "signer-1"},
					{IbcClientId: "client-1", Height: 2, Signer: "signer-1"},
				},
			},
			valid: true,
		},
		{
			name: "invalid rollapp id",
			g: types.GenesisState{
				CanonicalClients: []*types.CanonicalClient{
					{RollappId: "", IbcClientId: "client-1"},
				},
			},
			valid: false,
		},
		{
			name: "invalid ibc client id",
			g: types.GenesisState{
				CanonicalClients: []*types.CanonicalClient{
					{RollappId: "rollapp-1", IbcClientId: ""},
				},
			},
			valid: false,
		},
		{
			name: "invalid height",
			g: types.GenesisState{
				ConsensusStateSigners: []*types.ConsensusStateSigner{
					{IbcClientId: "client-1", Height: 0, Signer: "signer-1"},
				},
			},
			valid: false,
		},
		{
			name: "invalid signer",
			g: types.GenesisState{
				ConsensusStateSigners: []*types.ConsensusStateSigner{
					{IbcClientId: "client-1", Height: 1, Signer: ""},
				},
			},
			valid: false,
		},
		{
			name:  "empty",
			g:     types.GenesisState{},
			valid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				require.NoError(t, tt.g.Validate())
			} else {
				require.Error(t, tt.g.Validate())
			}
		})
	}
}
