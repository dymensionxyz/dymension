package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgCancelSellOrder_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		assetId         string
		assetType       AssetType
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid",
			assetId:   "my-name",
			assetType: TypeName,
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid",
			assetId:   "alias",
			assetType: TypeAlias,
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - (Name) not allow empty name",
			assetId:         "",
			assetType:       TypeName,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) not allow empty alias",
			assetId:         "",
			assetType:       TypeAlias,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) not allow invalid name",
			assetId:         "-my-name",
			assetType:       TypeName,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) not allow invalid alias",
			assetId:         "bad-alias",
			assetType:       TypeAlias,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - invalid owner",
			assetId:         "my-name",
			assetType:       TypeName,
			owner:           "dym1fl48vsnmsdzcv85q5",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - missing owner",
			assetId:         "my-name",
			assetType:       TypeName,
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - owner must be dym1",
			assetId:         "my-name",
			assetType:       TypeName,
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - not supported asset type",
			assetId:         "asset",
			assetType:       AssetType_AT_UNKNOWN,
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid asset type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgCancelSellOrder{
				AssetId:   tt.assetId,
				AssetType: tt.assetType,
				Owner:     tt.owner,
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
