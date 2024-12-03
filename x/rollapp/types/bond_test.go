package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/require"
)

func TestIsUpdateMinSeqBond(t *testing.T) {
	tests := []struct {
		name string
		coin sdk.Coin
		want bool
	}{
		{"valid coin", sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(100)), true},
		{"zero amount", sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(0)), false},
		{"wrong denom", sdk.NewCoin("wrongdenom", sdk.NewInt(100)), false},
		{"empty denom", sdk.Coin{Denom: "", Amount: sdk.NewInt(100)}, false},
		{"zero type", sdk.Coin{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUpdateMinSeqBond(tt.coin)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidateBasicMinSeqBond(t *testing.T) {
	tests := []struct {
		name    string
		coin    sdk.Coin
		wantErr bool
	}{
		{"valid coin", sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(100)), false},
		{"zero amount", sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(0)), true},
		{"wrong denom", sdk.NewCoin("wrongdenom", sdk.NewInt(100)), true},
		{"empty denom", sdk.Coin{Denom: "", Amount: sdk.NewInt(100)}, true},
		{"zero type", sdk.Coin{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBasicMinSeqBond(tt.coin)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBasicMinSeqBondCoins(t *testing.T) {
	tests := []struct {
		name    string
		coins   sdk.Coins
		wantErr bool
	}{
		{"valid coins", sdk.NewCoins(sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(100))), false},
		{"zero amount", sdk.NewCoins(sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(0))), true},
		{"wrong denom", sdk.NewCoins(sdk.NewCoin("wrongdenom", sdk.NewInt(100))), true},
		{"empty denom", sdk.Coins{sdk.Coin{Denom: "", Amount: sdk.NewInt(100)}}, true},
		{"multiple coins", sdk.NewCoins(
			sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(100)),
			sdk.NewCoin("anotherdenom", sdk.NewInt(50)),
		), true},
		{"zero type", sdk.Coins{}, true},
		{"zero type alt", sdk.Coins{sdk.Coin{
			Denom:  "",
			Amount: math.Int{},
		}}, true},
		{"nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBasicMinSeqBondCoins(tt.coins)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
