package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/stretchr/testify/require"
)

func TestMsgPurchaseOrder_ValidateBasic(t *testing.T) {
	validOffer := testCoin(100)

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		assetId         string
		assetType       AssetType
		params          []string
		offer           sdk.Coin
		buyer           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid",
			assetId:   "my-name",
			assetType: TypeName,
			offer:     validOffer,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid",
			assetId:   "alias",
			assetType: TypeAlias,
			params:    []string{"rollapp_1-1"},
			offer:     validOffer,
			buyer:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - (Name) reject empty name",
			assetId:         "",
			assetType:       TypeName,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) reject empty alias",
			assetId:         "",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) bad name",
			assetId:         "-my-name",
			assetType:       TypeName,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) bad alias",
			assetId:         "bad-alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) reject non-empty params",
			assetId:         "my-name",
			assetType:       TypeName,
			params:          []string{"one"},
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "not accept order params for asset type",
		},
		{
			name:            "fail - (Alias) reject empty params",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          nil,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "expect 1 order param of RollApp ID for asset type",
		},
		{
			name:            "fail - (Alias) reject bad params",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"-not-chain-id-"},
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid RollApp ID format",
		},
		{
			name:            "fail - (Name) missing offer",
			assetId:         "my-name",
			assetType:       TypeName,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - (Alias) missing offer",
			assetId:         "alias",
			params:          []string{"rollapp_1-1"},
			assetType:       TypeAlias,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid offer",
		},
		{
			name:            "fail - (Name) offer can not be zero",
			assetId:         "my-name",
			assetType:       TypeName,
			offer:           testCoin(0),
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:            "fail - (Alias) offer can not be zero",
			assetId:         "alias",
			assetType:       TypeAlias,
			params:          []string{"rollapp_1-1"},
			offer:           testCoin(0),
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "offer must be positive",
		},
		{
			name:      "fail - offer can not be negative",
			assetId:   "my-name",
			assetType: TypeName,
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
			assetId:         "my-name",
			assetType:       TypeName,
			offer:           validOffer,
			buyer:           "",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid buyer",
			assetId:         "my-name",
			assetType:       TypeName,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - buyer must be dym1",
			assetId:         "my-name",
			assetType:       TypeName,
			offer:           validOffer,
			buyer:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "buyer is not a valid bech32 account address",
		},
		{
			name:            "fail - reject unknown asset type",
			assetId:         "asset",
			assetType:       AssetType_AT_UNKNOWN,
			offer:           validOffer,
			buyer:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid asset type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPurchaseOrder{
				AssetId:   tt.assetId,
				AssetType: tt.assetType,
				Params:    tt.params,
				Offer:     tt.offer,
				Buyer:     tt.buyer,
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
