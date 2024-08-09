package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrateChainIdsProposal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		title           string
		description     string
		replacement     []MigrateChainId
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "pass - valid, single",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr: false,
		},
		{
			name:        "pass - valid, multiple",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "columbus-4",
					NewChainId:      "columbus-5",
				},
			},
			wantErr: false,
		},
		{
			name:            "fail - reject empty replacement",
			title:           "T",
			description:     "D",
			replacement:     []MigrateChainId{},
			wantErr:         true,
			wantErrContains: "replacement cannot be empty",
		},
		{
			name:            "fail - reject empty replacement",
			title:           "T",
			description:     "D",
			replacement:     nil,
			wantErr:         true,
			wantErrContains: "replacement cannot be empty",
		},
		{
			name:        "fail - reject empty title",
			title:       "",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr:         true,
			wantErrContains: "proposal title cannot be blank",
		},
		{
			name:        "fail - reject empty description",
			title:       "T",
			description: "",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr:         true,
			wantErrContains: "proposal description cannot be blank",
		},
		{
			name:        "fail - reject invalid replacement",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr:         true,
			wantErrContains: "previous chain id cannot be empty",
		},
		{
			name:        "fail - reject duplicate replacement",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-5",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id",
		},
		{
			name:        "fail - reject duplicate replacement",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "cosmoshub-2",
					NewChainId:      "cosmoshub-4",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id",
		},
		{
			name:        "fail - reject duplicate replacement",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "cosmoshub-2",
					NewChainId:      "cosmoshub-3",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id",
		},
		{
			name:        "fail - reject duplicate replacement",
			title:       "T",
			description: "D",
			replacement: []MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "cosmoshub-4",
					NewChainId:      "cosmoshub-5",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MigrateChainIdsProposal{
				Title:       tt.title,
				Description: tt.description,
				Replacement: tt.replacement,
			}

			err := m.ValidateBasic()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMigrateChainId_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		previousChainId string
		newChainId      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "pass - valid",
			previousChainId: "cosmoshub-3",
			newChainId:      "cosmoshub-4",
			wantErr:         false,
		},
		{
			name:            "fail - not allow empty previous chain-id",
			previousChainId: "",
			newChainId:      "cosmoshub-4",
			wantErr:         true,
			wantErrContains: "previous chain id cannot be empty",
		},
		{
			name:            "fail - not allow empty new chain-id",
			previousChainId: "cosmoshub-3",
			newChainId:      "",
			wantErr:         true,
			wantErrContains: "new chain id cannot be empty",
		},
		{
			name:            "fail - chain-id cannot be the same",
			previousChainId: "cosmoshub-3",
			newChainId:      "cosmoshub-3",
			wantErr:         true,
			wantErrContains: "previous chain id and new chain id cannot be the same",
		},
		{
			name:            "fail - chain-id cannot be the same, case insensitive",
			previousChainId: "CosmosHub-3",
			newChainId:      "cosmoshub-3",
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
		{
			name:            "fail - reject invalid previous chain-id",
			previousChainId: "cosmoshub@3",
			newChainId:      "cosmoshub-4",
			wantErr:         true,
			wantErrContains: "previous chain id is not well-formed:",
		},
		{
			name:            "fail - reject invalid new chain-id",
			previousChainId: "cosmoshub-3",
			newChainId:      "cosmoshub@4",
			wantErr:         true,
			wantErrContains: "new chain id is not well-formed:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MigrateChainId{
				PreviousChainId: tt.previousChainId,
				NewChainId:      tt.newChainId,
			}

			err := m.ValidateBasic()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
