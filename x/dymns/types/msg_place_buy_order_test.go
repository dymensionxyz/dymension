package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestMsgPlaceBuyOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name            string
		assetId         string
		assetType       AssetType
		params          []string
		buyer           string
		continueOrderId string
		offer           sdk.Coin
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "pass - (Name) valid",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOrderId: "",
			offer:           testCoin(1),
			wantErr:         false,
		},
		{
			name:            "pass - (Alias) valid",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOrderId: "",
			offer:           testCoin(1),
			wantErr:         false,
		},
		{
			name:            "pass - (Name) valid, continue offer",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOrderId: "101",
			offer:           testCoin(1),
			wantErr:         false,
		},
		{
			name:            "pass - (Alias) valid, continue offer",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOrderId: "101",
			offer:           testCoin(1),
			wantErr:         false,
		},
		{
			name:            "fail - (Name) reject bad Dym-Name format",
			assetId:         "@",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) reject bad alias format",
			assetId:         "bad-alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) reject bad params",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          []string{"not-empty"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "not accept order params for asset type",
		},
		{
			name:            "fail - (Alias) reject empty params",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          nil,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for asset type",
		},
		{
			name:            "fail - (Alias) reject bad params",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"@chain-id"},
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format",
		},
		{
			name:            "fail - bad buyer",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - offer ID",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			continueOrderId: "@",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "continue offer id is not a valid offer id",
		},
		{
			name:            "fail - empty offer",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           sdk.Coin{},
			wantErr:         true,
			wantErrContains: "invalid offer amount",
		},
		{
			name:            "fail - zero offer",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(0),
			wantErr:         true,
			wantErrContains: "offer amount must be positive",
		},
		{
			name:      "fail - negative offer",
			assetId:   "my-name",
			assetType: TypeName,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			wantErr:         true,
			wantErrContains: "invalid offer amount",
		},
		{
			name:            "fail - reject unknown asset type",
			assetId:         "asset",
			assetType:       AssetType_AT_UNKNOWN,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			offer:           testCoin(1),
			wantErr:         true,
			wantErrContains: "invalid asset type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPlaceBuyOrder{
				AssetId:         tt.assetId,
				AssetType:       tt.assetType,
				Params:          tt.params,
				Buyer:           tt.buyer,
				ContinueOrderId: tt.continueOrderId,
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
