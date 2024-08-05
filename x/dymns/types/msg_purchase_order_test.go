package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestMsgPurchaseOrder_ValidateBasic(t *testing.T) {
	validOffer := dymnsutils.TestCoin(100)

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		offer           sdk.Coin
		buyer           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "pass - valid",
			dymName: "abc",
			offer:   validOffer,
			buyer:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - not allow missing name",
			dymName:         "",
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - bad name",
			dymName:         "-a",
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - missing offer",
			dymName:         "abc",
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - offer can not be zero",
			dymName:         "abc",
			offer:           dymnsutils.TestCoin(0),
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:    "fail - offer can not be negative",
			dymName: "abc",
			offer: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - missing buyer",
			dymName:         "abc",
			offer:           validOffer,
			buyer:           "",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid buyer",
			dymName:         "abc",
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - buyer must be dym1",
			dymName:         "abc",
			offer:           validOffer,
			buyer:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPurchaseOrder{
				Name:  tt.dymName,
				Offer: tt.offer,
				Buyer: tt.buyer,
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
