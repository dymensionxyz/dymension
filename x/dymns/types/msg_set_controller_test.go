package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgSetController_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		controller      string
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "pass - valid",
			dymName:    "a",
			controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "pass - controller and owner can be the same",
			dymName:    "a",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:       "pass - controller and owner can be the different",
			dymName:    "a",
			controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - missing name",
			dymName:         "",
			controller:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - missing controller",
			dymName:         "a",
			controller:      "",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid controller",
			dymName:         "a",
			controller:      "dym1tygms3xhhs3yv487phx",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
		{
			name:            "fail - missing owner",
			dymName:         "a",
			controller:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid owner",
			dymName:         "a",
			controller:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "dym1fl48vsnmsdzcv85q5d2",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - controller must be dym1",
			dymName:         "a",
			controller:      "nim1tygms3xhhs3yv487phx3dw4a95jn7t7l4kreyj",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
		{
			name:            "fail - owner must be dym1",
			dymName:         "a",
			controller:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgSetController{
				Name:       tt.dymName,
				Controller: tt.controller,
				Owner:      tt.owner,
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
