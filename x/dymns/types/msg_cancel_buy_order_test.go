package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestMsgCancelBuyOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		orderId         string
		buyer           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "pass - valid",
			orderId: "101",
			buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr: false,
		},
		{
			name:            "fail - bad offer id",
			orderId:         "@",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer id is not a valid buy name offer id",
		},
		{
			name:            "fail - bad buyer",
			orderId:         "101",
			buyer:           "dym1fl48vsnmsdzcv85",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgCancelBuyOrder{
				OrderId: tt.orderId,
				Buyer:   tt.buyer,
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
