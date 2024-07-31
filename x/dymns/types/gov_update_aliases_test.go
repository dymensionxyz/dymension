package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateAliasesProposal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		title           string
		description     string
		add             []UpdateAlias
		remove          []UpdateAlias
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "valid, single add",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple add",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple add of same chain id",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dymension",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, single remove",
			title:       "T",
			description: "D",
			remove: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple remove",
			title:       "T",
			description: "D",
			remove: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple remove of same chain id",
			title:       "T",
			description: "D",
			remove: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dymension",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple add and remove",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
			},
			remove: []UpdateAlias{
				{
					ChainId: "froopyland_111-1",
					Alias:   "fl",
				},
			},
			wantErr: false,
		},
		{
			name:        "valid, multiple add and remove of same chain id",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym2",
				},
			},
			remove: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym3",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym4",
				},
			},
			wantErr: false,
		},
		{
			name:        "reject empty title",
			title:       "",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr:         true,
			wantErrContains: "proposal title cannot be blank",
		},
		{
			name:        "reject empty description",
			title:       "T",
			description: "",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr:         true,
			wantErrContains: "proposal description cannot be blank",
		},
		{
			name:            "reject empty proposal",
			title:           "T",
			description:     "D",
			wantErr:         true,
			wantErrContains: "update list can not be empty",
		},
		{
			name:        "reject non-unique combination of chain id and alias",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
			},
			remove: []UpdateAlias{
				{
					ChainId: "froopyland_111-1",
					Alias:   "fl",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id and alias pair",
		},
		{
			name:        "reject non-unique combination of chain id and alias",
			title:       "T",
			description: "D",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "blumbus",
				},
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym1",
				},
			},
			remove: []UpdateAlias{
				{
					ChainId: "froopyland_111-1",
					Alias:   "fl",
				},
			},
			wantErr:         true,
			wantErrContains: "duplicate chain id and alias pair",
		},
		{
			name:        "reject record that does not pass validation",
			title:       "T",
			description: "",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "Blumbus_111-1", // bad chain id
					Alias:   "bb",
				},
			},
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := UpdateAliasesProposal{
				Title:       tt.title,
				Description: tt.description,
				Add:         tt.add,
				Remove:      tt.remove,
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

func TestUpdateAlias_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		chainId         string
		alias           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "valid",
			chainId: "dymension_1100-1",
			alias:   "dym",
			wantErr: false,
		},
		{
			name:            "chain-id can not be empty",
			chainId:         "",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id cannot be empty",
		},
		{
			name:            "chain-id must be well-formed",
			chainId:         "@dym",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
		{
			name:            "chain-id must be well-formed",
			chainId:         "Dymension_1100-1",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
		{
			name:            "alias can not be empty",
			chainId:         "dymension_1100-1",
			alias:           "",
			wantErr:         true,
			wantErrContains: "alias cannot be empty",
		},
		{
			name:            "alias must be well-formed",
			chainId:         "dymension_1100-1",
			alias:           "@dym",
			wantErr:         true,
			wantErrContains: "alias is not well-formed",
		},
		{
			name:            "alias must be well-formed",
			chainId:         "dymension_1100-1",
			alias:           "Dym",
			wantErr:         true,
			wantErrContains: "alias is not well-formed",
		},
		{
			name:            "chain-id and alias can not be the same",
			chainId:         "dymension",
			alias:           "dymension",
			wantErr:         true,
			wantErrContains: "chain id and alias cannot be the same",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &UpdateAlias{
				ChainId: tt.chainId,
				Alias:   tt.alias,
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
