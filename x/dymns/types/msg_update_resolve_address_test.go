package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgUpdateResolveAddress_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		chainId         string
		subName         string
		resolveTo       string
		controller      string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "valid",
			dymName:    "a",
			chainId:    "dymension_1100-1",
			subName:    "abc",
			resolveTo:  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "missing dym-name",
			dymName:         "",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "bad dym-name",
			dymName:         "",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:       "valid config resolve with multi-level sub-name",
			dymName:    "a",
			subName:    "abc.def",
			resolveTo:  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "valid config resolve without sub-name",
			dymName:    "a",
			chainId:    "dymension_1100-1",
			subName:    "",
			resolveTo:  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "valid config resolve with empty chain-id",
			dymName:    "a",
			chainId:    "",
			subName:    "abc",
			resolveTo:  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "valid config resolve with empty chain-id and sub-name",
			dymName:    "a",
			chainId:    "",
			subName:    "",
			resolveTo:  "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "bad chain-id",
			dymName:         "a",
			chainId:         "dymension_",
			subName:         "abc",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config chain id must be a valid chain id format",
		},
		{
			name:            "bad sub-name",
			dymName:         "a",
			chainId:         "",
			subName:         "-a",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config path must be a valid dym name",
		},
		{
			name:            "bad sub-name, too long",
			dymName:         "a",
			chainId:         "",
			subName:         "123456789012345678901",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "sub name is too long",
		},
		{
			name:            "bad multi-level sub-name",
			dymName:         "a",
			chainId:         "",
			subName:         "a.b.",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config path must be a valid dym name",
		},
		{
			name:       "resolve to can be empty to allow delete",
			dymName:    "a",
			chainId:    "",
			subName:    "a",
			resolveTo:  "",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "resolve to can be empty to allow delete",
			dymName:    "a",
			chainId:    "",
			subName:    "",
			resolveTo:  "",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "bad resolve to",
			dymName:         "a",
			chainId:         "",
			subName:         "a",
			resolveTo:       "0x01",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config value must be a valid bech32 account address",
		},
		{
			name:            "resolve must be dym1 format if chain-id is empty",
			dymName:         "a",
			chainId:         "",
			resolveTo:       "nim1tygms3xhhs3yv487phx3dw4a95jn7t7l4kreyj",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "resolve address must be a valid bech32 account address on host chain",
		},
		{
			name:       "resolve to can be non-dym1 format if chain-id is not empty",
			dymName:    "a",
			chainId:    "nim_1122-1",
			resolveTo:  "nim1tygms3xhhs3yv487phx3dw4a95jn7t7l4kreyj",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "controller must be dym1",
			dymName:         "a",
			resolveTo:       "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "nim1tygms3xhhs3yv487phx3dw4a95jn7t7l4kreyj",
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgUpdateResolveAddress{
				Name:       tt.dymName,
				ChainId:    tt.chainId,
				SubName:    tt.subName,
				ResolveTo:  tt.resolveTo,
				Controller: tt.controller,
			}

			err := m.ValidateBasic()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateResolveAddress_GetDymNameConfig(t *testing.T) {
	tests := []struct {
		name       string
		DymName    string
		ChainId    string
		SubName    string
		ResolveTo  string
		Controller string
		wantName   string
		wantConfig DymNameConfig
	}{
		{
			name:       "assigned correctly",
			DymName:    "a",
			ChainId:    "dymension",
			SubName:    "sub",
			ResolveTo:  "r",
			Controller: "c",
			wantName:   "a",
			wantConfig: DymNameConfig{
				Type:    DymNameConfigType_NAME,
				ChainId: "dymension",
				Path:    "sub",
				Value:   "r",
			},
		},
		{
			name: "all empty",
			wantConfig: DymNameConfig{
				Type: DymNameConfigType_NAME,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgUpdateResolveAddress{
				Name:       tt.DymName,
				ChainId:    tt.ChainId,
				SubName:    tt.SubName,
				ResolveTo:  tt.ResolveTo,
				Controller: tt.Controller,
			}

			gotName, gotConfig := m.GetDymNameConfig()
			require.Equal(t, tt.wantName, gotName)
			require.Equal(t, tt.wantConfig, gotConfig)
		})
	}
}
