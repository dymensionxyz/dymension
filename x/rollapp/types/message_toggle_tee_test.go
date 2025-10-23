package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgToggleTEE_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgToggleTEE
		wantErr bool
	}{
		{
			name: "valid - enable TEE",
			msg: MsgToggleTEE{
				Owner:     sample.AccAddress(),
				RollappId: "dym_100-1",
				Enable:    true,
			},
			wantErr: false,
		},
		{
			name: "valid - disable TEE",
			msg: MsgToggleTEE{
				Owner:     sample.AccAddress(),
				RollappId: "dym_100-1",
				Enable:    false,
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgToggleTEE{
				Owner:     "invalid_address",
				RollappId: "dym_100-1",
				Enable:    true,
			},
			wantErr: true,
		},
		{
			name: "empty owner address",
			msg: MsgToggleTEE{
				Owner:     "",
				RollappId: "dym_100-1",
				Enable:    true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err, "test %s failed", tt.name)
				return
			}
			require.NoError(t, err)
		})
	}
}
