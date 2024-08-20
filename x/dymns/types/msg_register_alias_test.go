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
func TestMsgRegisterAlias_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		alias           string
		rollAppId       string
		owner           string
		confirmPayment  sdk.Coin
		wantErr         bool
		wantErrContains string
	}{
		{
			name:           "pass - valid",
			alias:          "a",
			rollAppId:      "rollapp_1-1",
			owner:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: testCoin(1),
		},
		{
			name:            "fail - missing alias",
			alias:           "",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "alias is not a valid alias format",
		},
		{
			name:            "fail - alias is too long",
			alias:           "123456789012345678901234567890123",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("alias is too long, maximum %d characters", dymnsutils.MaxAliasLength),
		},
		{
			name:            "fail - invalid alias",
			alias:           "-a",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "alias is not a valid alias format",
		},
		{
			name:            "fail - empty RollApp ID",
			alias:           "a",
			rollAppId:       "",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "RollApp ID is not a valid chain id format",
		},
		{
			name:            "fail - bad RollApp ID",
			alias:           "a",
			rollAppId:       "-RollApp",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "RollApp ID is not a valid chain id format",
		},
		{
			name:            "fail - empty owner",
			alias:           "a",
			rollAppId:       "rollapp_1-1",
			owner:           "",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid owner",
			alias:           "a",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - owner must be dym1",
			alias:           "a",
			rollAppId:       "rollapp_1-1",
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			confirmPayment:  testCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - confirm-payment not set",
			alias:           "a",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  sdk.Coin{},
			wantErr:         true,
			wantErrContains: "confirm payment is not set",
		},
		{
			name:            "fail - confirm-payment is zero",
			alias:           "a",
			rollAppId:       "rollapp_1-1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment:  testCoin(0),
			wantErr:         true,
			wantErrContains: "confirm payment is not set",
		},
		{
			name:      "fail - confirm-payment is negative",
			alias:     "a",
			rollAppId: "rollapp_1-1",
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdkmath.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "negative coin amount",
		},
		{
			name:      "fail - confirm-payment without denom",
			alias:     "a",
			rollAppId: "rollapp_1-1",
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			confirmPayment: sdk.Coin{
				Denom:  "",
				Amount: sdk.OneInt(),
			},
			wantErr:         true,
			wantErrContains: "invalid denom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgRegisterAlias{
				Alias:          tt.alias,
				RollappId:      tt.rollAppId,
				Owner:          tt.owner,
				ConfirmPayment: tt.confirmPayment,
			}

			err := m.ValidateBasic()
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
