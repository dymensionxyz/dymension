package types

import (
	"testing"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/stretchr/testify/require"
)

func TestMsgPlaceSellOrder_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		assetId         string
		assetType       AssetType
		minPrice        sdk.Coin
		sellPrice       *sdk.Coin
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "pass - (Name) valid sell order",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  testCoin(1),
			sellPrice: uptr.To(testCoin(1)),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid sell order",
			assetId:   "alias",
			assetType: TypeAlias,
			minPrice:  testCoin(1),
			sellPrice: uptr.To(testCoin(1)),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Name) valid sell order without bid",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  testCoin(1),
			sellPrice: uptr.To(testCoin(1)),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid sell order without bid",
			assetId:   "alias",
			assetType: TypeAlias,
			minPrice:  testCoin(1),
			sellPrice: uptr.To(testCoin(1)),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Name) valid sell order without setting sell price",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  testCoin(1),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "pass - (Alias) valid sell order without setting sell price",
			assetId:   "alias",
			assetType: TypeAlias,
			minPrice:  testCoin(1),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "fail - (Name) empty name",
			assetId:         "",
			assetType:       TypeName,
			minPrice:        testCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) empty alias",
			assetId:         "",
			assetType:       TypeAlias,
			minPrice:        testCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - (Name) bad name",
			assetId:         "-my-name",
			assetType:       TypeName,
			minPrice:        testCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "fail - (Alias) bad alias",
			assetId:         "bad-alias",
			assetType:       TypeAlias,
			minPrice:        testCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "alias is not a valid alias",
		},
		{
			name:            "fail - min price is zero",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(0),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "fail - min price is empty",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        sdk.Coin{},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:      "fail - min price is negative",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is negative",
		},
		{
			name:      "fail - min price is invalid",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is invalid",
		},
		{
			name:            "fail - sell price is negative",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(1),
			sellPrice:       uptr.To(testCoin(-1)),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is negative",
		},
		{
			name:      "fail - sell price is invalid",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  testCoin(1),
			sellPrice: &sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is invalid",
		},
		{
			name:            "fail - sell price is less than min price",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(2),
			sellPrice:       uptr.To(testCoin(1)),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is less than min price",
		},
		{
			name:            "fail - sell price denom must match min price denom",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(2),
			sellPrice:       uptr.To(sdk.NewCoin("u"+params.BaseDenom, sdk.OneInt())),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price denom is different from min price denom",
		},
		{
			name:            "fail - missing owner",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(2),
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - invalid owner",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(2),
			owner:           "dym1fl48vsnmsdzcv85",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - owner must be dym1",
			assetId:         "my-name",
			assetType:       TypeName,
			minPrice:        testCoin(2),
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "fail - reject unknown asset type",
			assetId:         "asset",
			assetType:       AssetType_AT_UNKNOWN,
			minPrice:        testCoin(2),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "invalid asset type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPlaceSellOrder{
				AssetId:   tt.assetId,
				AssetType: tt.assetType,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
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

func TestMsgPlaceSellOrder_ToSellOrder(t *testing.T) {
	validMinPrice := testCoin(1)
	validSellPrice := testCoin(1)

	tests := []struct {
		name      string
		assetId   string
		assetType AssetType
		minPrice  sdk.Coin
		sellPrice *sdk.Coin
		Owner     string
		want      SellOrder
	}{
		{
			name:      "normal Dym-Name sell order",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  validMinPrice,
			sellPrice: &validSellPrice,
			Owner:     "",
			want: SellOrder{
				AssetId:   "my-name",
				AssetType: TypeName,
				MinPrice:  validMinPrice,
				SellPrice: &validSellPrice,
			},
		},
		{
			name:      "normal Alias sell order",
			assetId:   "alias",
			assetType: TypeAlias,
			minPrice:  validMinPrice,
			sellPrice: &validSellPrice,
			Owner:     "",
			want: SellOrder{
				AssetId:   "alias",
				AssetType: TypeAlias,
				MinPrice:  validMinPrice,
				SellPrice: &validSellPrice,
			},
		},
		{
			name:      "without sell price",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  validMinPrice,
			sellPrice: nil,
			Owner:     "",
			want: SellOrder{
				AssetId:   "my-name",
				AssetType: TypeName,
				MinPrice:  validMinPrice,
				SellPrice: nil,
			},
		},
		{
			name:      "without sell price, auto omit zero sell price",
			assetId:   "my-name",
			assetType: TypeName,
			minPrice:  validMinPrice,
			sellPrice: uptr.To(sdk.NewCoin(validMinPrice.Denom, sdk.ZeroInt())),
			Owner:     "",
			want: SellOrder{
				AssetId:   "my-name",
				AssetType: TypeName,
				MinPrice:  validMinPrice,
				SellPrice: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPlaceSellOrder{
				AssetId:   tt.assetId,
				AssetType: tt.assetType,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
				Owner:     tt.Owner,
			}

			so := m.ToSellOrder()
			require.Equal(t, tt.want, so)
			require.Zero(t, so.ExpireAt)
			require.Nil(t, so.HighestBid)
		})
	}
}
