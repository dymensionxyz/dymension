package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestMsgRegisterName_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		dymName         string
		duration        int64
		owner           string
		confirmPayment  sdk.Coin
		contact         string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:           "valid 1 yr",
			dymName:        "a",
			duration:       1,
			owner:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: dymnsutils.TestCoin(1),
			contact:        "contact@example.com",
		},
		{
			name:           "valid 1+ yrs",
			dymName:        "a",
			duration:       5,
			owner:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: dymnsutils.TestCoin(1),
		},
		{
			name:            "missing name",
			dymName:         "",
			duration:        5,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "name is too long",
			dymName:         "123456789012345678901",
			duration:        5,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("name is too long, maximum %d characters", MaxDymNameLength),
		},
		{
			name:            "invalid name",
			dymName:         "-a",
			duration:        5,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "zero duration",
			dymName:         "a",
			duration:        0,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "duration must be at least 1 year",
		},
		{
			name:            "negative duration",
			dymName:         "a",
			duration:        -1,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "duration must be at least 1 year",
		},
		{
			name:            "empty owner",
			dymName:         "a",
			duration:        1,
			owner:           "",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "invalid owner",
			dymName:         "a",
			duration:        1,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "owner must be dym1",
			dymName:         "a",
			duration:        1,
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			confirmPayment:  dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "confirm-payment not set",
			dymName:         "a",
			duration:        1,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  sdk.Coin{},
			wantErr:         true,
			wantErrContains: "confirm payment is not set",
		},
		{
			name:            "confirm-payment is zero",
			dymName:         "a",
			duration:        1,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "confirm payment is not set",
		},
		{
			name:     "confirm-payment is negative",
			dymName:  "a",
			duration: 1,
			owner:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdkmath.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "negative coin amount",
		},
		{
			name:     "confirm-payment without denom",
			dymName:  "a",
			duration: 1,
			owner:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: sdk.Coin{
				Denom:  "",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "invalid denom",
		},
		{
			name:            "contact too long",
			dymName:         "a",
			duration:        1,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  dymnsutils.TestCoin(1),
			contact:         "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			wantErr:         true,
			wantErrContains: "invalid contact length",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgRegisterName{
				Name:           tt.dymName,
				Duration:       tt.duration,
				Owner:          tt.owner,
				ConfirmPayment: tt.confirmPayment,
				Contact:        tt.contact,
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
