package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestMsgPutAdsSellName_ValidateBasic(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		dymName         string
		minPrice        sdk.Coin
		sellPrice       *sdk.Coin
		owner           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:      "valid sell order",
			dymName:   "bonded-pool",
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:      "valid sell order without bid",
			dymName:   "bonded-pool",
			minPrice:  dymnsutils.TestCoin(1),
			sellPrice: dymnsutils.TestCoinP(1),
			owner:     "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:     "valid sell order without setting sell price",
			dymName:  "bonded-pool",
			minPrice: dymnsutils.TestCoin(1),
			owner:    "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "empty name",
			dymName:         "",
			minPrice:        dymnsutils.TestCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "bad name",
			dymName:         "-bonded-pool",
			minPrice:        dymnsutils.TestCoin(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:            "min price is zero",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(0),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:            "min price is empty",
			dymName:         "bonded-pool",
			minPrice:        sdk.Coin{},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is zero",
		},
		{
			name:    "min price is negative",
			dymName: "bonded-pool",
			minPrice: sdk.Coin{
				Denom:  params.BaseDenom,
				Amount: sdk.NewInt(-1),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is negative",
		},
		{
			name:    "min price is invalid",
			dymName: "bonded-pool",
			minPrice: sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO min price is invalid",
		},
		{
			name:            "sell price is negative",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(1),
			sellPrice:       dymnsutils.TestCoinP(-1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is negative",
		},
		{
			name:     "sell price is invalid",
			dymName:  "bonded-pool",
			minPrice: dymnsutils.TestCoin(1),
			sellPrice: &sdk.Coin{
				Denom:  "-",
				Amount: sdk.OneInt(),
			},
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is invalid",
		},
		{
			name:            "sell price is less than min price",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(2),
			sellPrice:       dymnsutils.TestCoinP(1),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price is less than min price",
		},
		{
			name:            "sell price denom must match min price denom",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(2),
			sellPrice:       dymnsutils.TestCoin2P(sdk.NewCoin("u"+params.BaseDenom, sdk.OneInt())),
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "SO sell price denom is different from min price denom",
		},
		{
			name:            "missing owner",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(2),
			owner:           "",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "invalid owner",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(2),
			owner:           "dym1fl48vsnmsdzcv85",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:            "owner must be dym1",
			dymName:         "bonded-pool",
			minPrice:        dymnsutils.TestCoin(2),
			owner:           "nim1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3pklgjx",
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPutAdsSellName{
				Name:      tt.dymName,
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

func TestMsgPutAdsSellName_ToSellOrder(t *testing.T) {
	validMinPrice := dymnsutils.TestCoin(1)
	validSellPrice := dymnsutils.TestCoin(1)

	tests := []struct {
		name      string
		Name      string
		MinPrice  sdk.Coin
		SellPrice *sdk.Coin
		Owner     string
		want      SellOrder
	}{
		{
			name:      "valid",
			Name:      "a",
			MinPrice:  validMinPrice,
			SellPrice: &validSellPrice,
			Owner:     "",
			want: SellOrder{
				Name:      "a",
				MinPrice:  validMinPrice,
				SellPrice: &validSellPrice,
			},
		},
		{
			name:      "valid without sell price",
			Name:      "a",
			MinPrice:  validMinPrice,
			SellPrice: nil,
			Owner:     "",
			want: SellOrder{
				Name:      "a",
				MinPrice:  validMinPrice,
				SellPrice: nil,
			},
		},
		{
			name:      "valid without sell price, auto omit zero sell price",
			Name:      "a",
			MinPrice:  validMinPrice,
			SellPrice: dymnsutils.TestCoin2P(sdk.NewCoin(validMinPrice.Denom, sdk.ZeroInt())),
			Owner:     "",
			want: SellOrder{
				Name:      "a",
				MinPrice:  validMinPrice,
				SellPrice: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgPutAdsSellName{
				Name:      tt.Name,
				MinPrice:  tt.MinPrice,
				SellPrice: tt.SellPrice,
				Owner:     tt.Owner,
			}

			so := m.ToSellOrder()
			require.Equal(t, tt.want, so)
			require.Zero(t, so.ExpireAt)
			require.Nil(t, so.HighestBid)
		})
	}
}
