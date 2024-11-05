package keeper_test

import (
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
)

func Test_proposerChoiceAlgo(t *testing.T) {
	type args struct {
		seqs []types.Sequencer
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "only one",
			args: args{
				seqs: []types.Sequencer{
					{
						Address: "0",
						Tokens:  sdk.NewCoins(bond),
					},
				},
			},
			want: 0,
		},
		{
			name: "two",
			args: args{
				seqs: []types.Sequencer{
					{
						Address: "0",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 1)),
					},
					{
						Address: "1",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 2)),
					},
				},
			},
			want: 1,
		},
		{
			name: "stable",
			args: args{
				seqs: []types.Sequencer{
					{
						Address: "0",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 1)),
					},
					{
						Address: "1",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 3)),
					},
					{
						Address: "2",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 3)),
					},
					{
						Address: "3",
						Tokens:  sdk.NewCoins(ucoin.SimpleMul(bond, 2)),
					},
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.args.seqs[tt.want]
			if got, _ := keeper.ProposerChoiceAlgo(tt.args.seqs); !reflect.DeepEqual(got, want) {
				t.Errorf("proposerChoiceAlgo() = %v, want %v", got, want)
			}
		})
	}
}
