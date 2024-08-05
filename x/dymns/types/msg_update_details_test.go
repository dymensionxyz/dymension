package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgUpdateDetails_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		controller      string
		contact         string
		clearConfigs    bool
		wantErr         bool
		wantErrContains string
	}{
		{
			name:         "pass - valid",
			dymName:      "a",
			controller:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:      "contact@example.com",
			clearConfigs: false,
		},
		{
			name:         "pass - valid",
			dymName:      "a",
			controller:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:      "",
			clearConfigs: false,
		},
		{
			name:         "pass - valid",
			dymName:      "a",
			controller:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:      "",
			clearConfigs: true,
		},
		{
			name:         "pass - valid",
			dymName:      "a",
			controller:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:      "contact@example.com",
			clearConfigs: true,
		},
		{
			name:            "fail - reject contact too long",
			dymName:         "a",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:         "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			clearConfigs:    true,
			wantErr:         true,
			wantErrContains: "contact is too long",
		},
		{
			name:            "fail - reject bad Dym-Name",
			dymName:         "a@",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:         "",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - reject bad controller",
			dymName:         "a",
			controller:      "dym1fl48vsnmsdzcv85q",
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
		{
			name:            "fail - reject message that does nothing",
			dymName:         "a",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			contact:         DoNotModifyDesc,
			clearConfigs:    false,
			wantErr:         true,
			wantErrContains: "message neither clears configs nor updates contact information",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgUpdateDetails{
				Name:         tt.dymName,
				Controller:   tt.controller,
				Contact:      tt.contact,
				ClearConfigs: tt.clearConfigs,
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
