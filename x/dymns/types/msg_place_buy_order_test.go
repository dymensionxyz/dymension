package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestMsgPlaceBuyOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		dymName         string
		buyer           string
		continueOfferId string
		offer           sdk.Coin
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "pass - valid",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOfferId: "",
			offer:           dymnsutils.TestCoin(1),
			wantErr:         false,
		},
		{
			name:            "pass - valid, continue offer",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOfferId: "101",
			offer:           dymnsutils.TestCoin(1),
			wantErr:         false,
		},
		{
			name:            "fail - bad Dym-Name",
			dymName:         "@",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - bad buyer",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85",
			offer:           dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - offer ID",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOfferId: "@",
			offer:           dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "continue offer id is not a valid buy name offer id",
		},
		{
			name:            "fail - empty offer",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           sdk.Coin{},
			wantErr:         true,
			wantErrContains: "invalid offer amount",
		},
		{
			name:            "fail - zero offer",
			dymName:         "a",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "offer amount must be positive",
		},
		{
			name:    "fail - negative offer",
			dymName: "a",
			buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "invalid offer amount",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPlaceBuyOrder{
				Name:            tt.dymName,
				Buyer:           tt.buyer,
				ContinueOfferId: tt.continueOfferId,
				Offer:           tt.offer,
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
