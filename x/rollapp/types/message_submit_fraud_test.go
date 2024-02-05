package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgSubmitFraud_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSubmitFraud
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgSubmitFraud{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgSubmitFraud{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
