package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestMsgAcceptOfferBuyName_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		offerId         string
		owner           string
		minAccept       sdk.Coin
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - valid",
			offerId:   "1",
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			minAccept: dymnsutils.TestCoin(1),
			wantErr:   false,
		},
		{
			name:            "fail - reject bad offer id",
			offerId:         "@",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			minAccept:       dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "offer id is not a valid buy name offer id",
		},
		{
			name:            "fail - reject bad owner",
			offerId:         "1",
			owner:           "x",
			minAccept:       dymnsutils.TestCoin(1),
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - reject empty coin",
			offerId:         "1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			minAccept:       sdk.Coin{},
			wantErr:         true,
			wantErrContains: "invalid min-accept amount",
		},
		{
			name:            "fail - reject zero coin",
			offerId:         "1",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			minAccept:       dymnsutils.TestCoin(0),
			wantErr:         true,
			wantErrContains: "min-accept amount must be positive",
		},
		{
			name:    "fail - reject negative coin",
			offerId: "1",
			owner:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			minAccept: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "invalid min-accept amount",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgAcceptOfferBuyName{
				OfferId:   tt.offerId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
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
