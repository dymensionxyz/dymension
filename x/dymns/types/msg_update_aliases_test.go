package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgUpdateAliases_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		add             []UpdateAlias
		remove          []UpdateAlias
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "pass - valid, single add",
			add: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr: false,
		},
		{
			name: "pass - valid, multiple add",
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
			name: "pass - valid, multiple add of same chain id",
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
			name: "pass - valid, single remove",
			remove: []UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
			wantErr: false,
		},
		{
			name: "pass - valid, multiple remove",
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
			name: "pass - valid, multiple remove of same chain id",
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
			name: "pass - valid, multiple add and remove",
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
			name: "pass - valid, multiple add and remove of same chain id",
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
			name: "fail - reject non-unique combination of chain id and alias",
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
			name: "fail - reject non-unique combination of chain id and alias",
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
			name: "fail - reject record that does not pass validation",
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
			m := MsgUpdateAliases{
				Add:    tt.add,
				Remove: tt.remove,
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
			name:    "pass - valid",
			chainId: "dymension_1100-1",
			alias:   "dym",
			wantErr: false,
		},
		{
			name:            "fail - chain-id can not be empty",
			chainId:         "",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id cannot be empty",
		},
		{
			name:            "fail - chain-id must be well-formed",
			chainId:         "@dym",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
		{
			name:            "fail - chain-id must be well-formed",
			chainId:         "Dymension_1100-1",
			alias:           "dym",
			wantErr:         true,
			wantErrContains: "chain id is not well-formed",
		},
		{
			name:            "fail - alias can not be empty",
			chainId:         "dymension_1100-1",
			alias:           "",
			wantErr:         true,
			wantErrContains: "alias cannot be empty",
		},
		{
			name:            "fail - alias must be well-formed",
			chainId:         "dymension_1100-1",
			alias:           "@dym",
			wantErr:         true,
			wantErrContains: "alias is not well-formed",
		},
		{
			name:            "fail - alias must be well-formed",
			chainId:         "dymension_1100-1",
			alias:           "Dym",
			wantErr:         true,
			wantErrContains: "alias is not well-formed",
		},
		{
			name:            "fail - chain-id and alias can not be the same",
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
