package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_checkIfConsecutiveAliasLengths(t *testing.T) {
	type args struct {
		aliasPricingTable map[string]sdk.Coin
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "consecutive alias lengths",
			args: args{
				aliasPricingTable: map[string]sdk.Coin{
					"1": sdk.NewCoin("dym", sdk.NewInt(1)),
					"2": sdk.NewCoin("dym", sdk.NewInt(2)),
					"3": sdk.NewCoin("dym", sdk.NewInt(3)),
					"4": sdk.NewCoin("dym", sdk.NewInt(4)),
				},
			},
			wantErr: false,
		}, {
			name: "non-consecutive alias lengths",
			args: args{
				aliasPricingTable: map[string]sdk.Coin{
					"1": sdk.NewCoin("dym", sdk.NewInt(1)),
					"3": sdk.NewCoin("dym", sdk.NewInt(3)),
					"4": sdk.NewCoin("dym", sdk.NewInt(4)),
				},
			},
			wantErr: true,
		}, {
			name: "missing first alias length",
			args: args{
				aliasPricingTable: map[string]sdk.Coin{
					"2": sdk.NewCoin("dym", sdk.NewInt(2)),
					"3": sdk.NewCoin("dym", sdk.NewInt(3)),
					"4": sdk.NewCoin("dym", sdk.NewInt(4)),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkIfConsecutiveAliasLengths(tt.args.aliasPricingTable); (err != nil) != tt.wantErr {
				t.Errorf("checkIfConsecutiveAliasLengths() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
