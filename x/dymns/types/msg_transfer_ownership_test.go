package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgTransferOwnership_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		newOwner        string
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "valid",
			dymName:  "a",
			newOwner: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "new owner and owner can not be the same",
			dymName:         "a",
			newOwner:        "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "new owner must be different from the current owner",
		},
		{
			name:            "missing name",
			dymName:         "",
			newOwner:        "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "missing new owner",
			dymName:         "a",
			newOwner:        "",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "new owner is not a valid bech32 account address",
		},
		{
			name:            "invalid new owner",
			dymName:         "a",
			newOwner:        "dym1tygms3xhhs3yv487phx",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "new owner is not a valid bech32 account address",
		},
		{
			name:            "missing owner",
			dymName:         "a",
			newOwner:        "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "invalid owner",
			dymName:         "a",
			newOwner:        "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "dym1fl48vsnmsdzcv85q5d2",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "new owner must be dym1",
			dymName:         "a",
			newOwner:        "nim1tygms3xhhs3yv487phx3dw4a95jn7t7l4kreyj",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "new owner is not a valid bech32 account address",
		},
		{
			name:            "owner must be dym1",
			dymName:         "a",
			newOwner:        "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgTransferOwnership{
				Name:     tt.dymName,
				NewOwner: tt.newOwner,
				Owner:    tt.owner,
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
